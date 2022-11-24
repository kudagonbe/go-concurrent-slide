class: center, middle

# 並行処理は怖くない！<br>Goで並行処理を書いてみよう

###.right[Hikaru Imamoto]

---

# 今日のアジェンダ

1. Goの並行処理について
1. GoroutineとWaitGroupを使って並行処理を実装する
1. Channelを使ってGoroutine間で処理を連携する
1. Goの並行処理の利用例

---

class: center, middle

# Goの並行処理について

---

# Goの並行処理について

- Go言語は並行処理を実装しやすいと言われています
- その理由は以下に集約されます
  - Goroutineという機構を使うと、新たな並行処理を生成・実行する実装が容易に書ける
  - Channelという仕組みを使うと、並行処理同士で値の送受信が容易に実装できる
- 今日はそれらの機構の具体的な実装方法を紹介して、Go言語の並行処理の便利さを実感してもらうことを目標としています

---

class: center, middle

# GoroutineとWaitGroupを使って<br>並行処理を実装する

---

# Goroutineとは

- Goroutine(ゴルーチン)とはGoのプロセス内で生成される軽量スレッド
- Goはメイン処理も含めてGoroutineで実行されるため、必ず1つ以上のGoroutineが存在している
- Goroutineから他のGoroutineを簡単に起動できるためGoのプロセス内でいくつものGoroutineを起動させることができる
- Goroutineの生成・実行は非同期で行われるため、生成元のGoroutine上で生成先のGoroutineの実行終了を待つ必要はない
- Goで複数の処理を並行に実行したい場合は、Goroutineを活用する

---

# 例) サイコロ関数の実装

### サイコロ関数を作成

`dice/dice.go`
- サイコロ関数(`RollDice`)
  - 1～6の整数のいずれかをランダムに生成して出力

```go
package dice

import (
    "fmt"
    "time"
    "math/rand"
)

func RollDice(index int) {
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(6) + 1
	fmt.Printf("%d: dice number is %d\n", index, r)
}
```

---

# 例) サイコロ関数の実装

### サイコロ関数を呼び出し

`main.go`
- サイコロ関数を10回呼び出すメイン処理
  - 単純に関数を呼んでいるだけなので、この実装のままだとGoroutineは生成されない

```go
package main

import "github.com/kudagonbe/go-concurrent-slide/dice"

func main() {
	// call RollDice 10 times
	for i := 0; i < 10; i++ {
		dice.RollDice(i)
	}
}
```

---

### 実行結果

1回ごとにランダムな値を生成して出力

```bash
$ go run main.go
0: dice number is 1
1: dice number is 1
2: dice number is 6
3: dice number is 3
4: dice number is 2
5: dice number is 4
6: dice number is 2
7: dice number is 2
8: dice number is 5
9: dice number is 2
```

---

# サイコロ関数をGoroutineで呼び出し

`main.go`
- 先ほどの`RollDice()`関数呼び出しに`go`というキーワードをつける
  - `go`キーワードを付けて関数を呼び出すとGoroutineが新たに生成される

```go
package main

import "github.com/kudagonbe/go-concurrent-slide/dice"

func main() {
  // call RollDice 10 times with goroutine
	for i := 0; i < 10; i++ {
*		go dice.RollDice(i)
	}
}
```

---

### 実行結果

- 何も出力されない
  - `go`キーワードで生成したGoroutineは非同期実行されている
  - つまり`for`文は各Goroutineとして実行されている関数の実行終了を待たずに回り続ける
  - 結果、各Goroutineの`fmt.Printf()`関数が実行される前にメインのGoroutineの実行が終了するため、結果が出力されない
  - メインのGoroutineにsleep処理を入れれば出力されるが、各Goroutineの実行終了を正確に検知したい

```bash
$ go run main.go
$
```

---

# `sync.WaitGroup`を使う

- 現在実行中のGoroutineの数を`sync.WaitGroup`で管理
  - 実行開始前に`wg.Add()`, 実行終了時に`wg.Done()`を実行
  - 実行中のGoroutineが無くなるまで`wg.Wait()`で処理をブロック

---

# `sync.WaitGroup`を使う

`main.go`
```go
package main

import (
	"sync"
	"github.com/kudagonbe/go-concurrent-slide/dice"
)

func main() {
*	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
*		wg.Add(1)
		go func(i int) {
*			defer wg.Done()
			dice.RollDice(i)
		}(i)
	}
*	wg.Wait()
}
```

---

### 実行結果

- Goroutineは非同期実行されるので、処理順序は保証されない

```bash
go run main.go
9: dice number is 6
0: dice number is 2
5: dice number is 3
8: dice number is 5
6: dice number is 1
7: dice number is 5
4: dice number is 3
1: dice number is 6
2: dice number is 1
3: dice number is 1
```

---

class: center, middle

# Channelを使って<br>Goroutine間で処理を連携する

---

# サイコロの平均値を出す

- 並行に実行されたサイコロ関数で計算された値を集計して、サイコロの個数とそれぞれの値の個数、全体の平均値を出力したい
- Goroutineは非同期実行されるため、処理結果を返却しても生成元のGoroutineで戻り値を受け取ることができない
- Goroutine同士の値の受け渡しには`Channel`という機構を使う

---

# Channelとは

- `Channel`(チャネル)とは、特定の型の値をGoroutine間で送受信できるデータ型
- 例えば整数型を表すChannelは `chan int` というデータ型になる

```go
ch := make(chan int, 2) //容量2の整数型Channel作成
ch <- 1 // 1つめの値をChannelに送信
ch <- 2 // 2つめの値をChannelに送信
v1 := <-ch // 1つめの値を受信(1が格納される)
v2 := <-ch // 2つめの値を受信(2が格納される)
```

---

### サイコロ関数を修正

`dice/dice.go`
- 新たなサイコロ関数(`RollDiceWithChannel`)を追加
  - 引数でintを格納するChannelを受け取る
  - ランダム生成した値をChannelに送信する

```go
package dice

import (
    "time"
    "math/rand"
)

*func RollDiceWithChannel(ch chan int) {
	rand.Seed(time.Now().UnixNano())
*	ch <- rand.Intn(6) + 1
}
```

---

### main処理を修正

- `ch chan int`を生成し、サイコロ関数に渡している

`main.go`
```go
package main

import (
	"fmt"

	"github.com/kudagonbe/go-concurrent-slide/dice"
)

func main() {
	var count int = 100000
	var total int
	values := map[int]int{1: 0, 2: 0, 3: 0, 4: 0, 5: 0, 6: 0}

*	ch := make(chan int, count)
	done := make(chan struct{})

	for i := 0; i < count; i++ {
*		go dice.RollDiceWithChannel(ch)
	}
  // 次ページへ続く
}
```

---

### main処理を修正

- Channelの受信を待ち受けて検知する`select`文でサイコロ関数から送信された値を受け取る
- `for`文で`select`文を無限ループさせて、想定の個数の値を受信し終わったら`for`文から抜ける
- 結果出力を出力

```go
func main() {
  //前ページからの続き
Loop:
	for {
		select {
*		case value := <-ch:
			total++
			values[value]++
			if total >= count {
				close(done)
			}
*		case <-done:
			close(ch)
			break Loop
		}
	}

*	outputResults(values, total)
}
```

---

### 結果出力関数を作成

`main.go`
```go
func outputResults(values map[int]int, total int) {
	fmt.Printf("total: %d\n", total)
	fmt.Printf("1: %d\n", values[1])
	fmt.Printf("2: %d\n", values[2])
	fmt.Printf("3: %d\n", values[3])
	fmt.Printf("4: %d\n", values[4])
	fmt.Printf("5: %d\n", values[5])
	fmt.Printf("6: %d\n", values[6])
	fmt.Printf("average: %f\n", avg(values, total))

}

func avg(values map[int]int, total int) float64 {
	var f float64
	for k, v := range values {
		f += float64(k) * float64(v)
	}
	return f / float64(total)
}
```

---

### 実行結果

- だいたい3.5前後に収束する

```bash
$ go run main.go
total: 100000
1: 16666
2: 16761
3: 16588
4: 16708
5: 16625
6: 16652
average: 3.498210
```

---

class: center, middle

# Goの並行処理の利用例

---

### 例1. Job管理機能を自前で実装

- 定期的に実行したい関数を非同期に実行する機構を実装可能
- `time.NewTicker(duration)`で定期的にChannelに値を送信する`time.Ticker`を生成して、そのタイミングでGoroutineを実行

```go
func AnyJob() {
  // 10分おきにChannelに値を送信するticker
* ticker := time.NewTicker(time.Minute * 10)
  for {
    select
*     case t := <-ticker.C: // tickerからのChannel送信を受信
        go anyFnc()
        ticker.Reset(j.d)
    }
  }
}
```

---

### 例2. イベントキューになっているChannelを監視してGoroutine実行

- 分散サービスではあるサービスで発生したイベントを監視して後続処理を実行する場面が多い
- イベント発生を監視するGoroutineがイベントを検知して、Channelにイベント内容を送信する
- 後続処理が必要な別のgoroutineがイベントを受信して後続処理を実行する
- Channelを介してデータをやり取りするGoroutine同士は疎結合を保ちつつ、ほぼリアルタイムで後続処理が実行できるというメリットがある

---

class: center, middle

# まとめ

---

class: center, middle

# 並行処理を制する者は<br>Goを制する！！<br>(たぶん)

---

class: center, middle

# ご清聴<br>ありがとうございました