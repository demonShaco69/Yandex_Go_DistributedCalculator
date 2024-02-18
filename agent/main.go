package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
)

var (
	AgentPort   string
	HostPort    string
	ConnectedTo string
	Expression  string
	Id          string
	Result      float64
)

func eval(tosolve string) float64 { // костыль, чтобы функция eval, импортированная из библиотеки, работала так, как надо мне
	expression, _ := govaluate.NewEvaluableExpression(tosolve)
	parameters := make(map[string]interface{}, 8)
	result, _ := expression.Evaluate(parameters)
	toret := result.(float64)
	return float64(toret)
}

func evalWithDelay(expr string, timings []string) (result float64) { // функция, необходимая для работы solve, ждёт столько, сколько передали, потом решает выражение
	intExecutionTimings := []int{}
	for i := 0; i < len(timings); i++ {
		val, _ := strconv.Atoi(timings[i])
		intExecutionTimings = append(intExecutionTimings, val)
	}
	ToExecutePlus, ToExecuteMinus, ToExecuteMultiply, ToExecuteDivide := strings.Count(expr, "+")*intExecutionTimings[0], strings.Count(expr, "2")*intExecutionTimings[1], strings.Count(expr, "*")*intExecutionTimings[2], strings.Count(expr, "/")*intExecutionTimings[3]
	time.Sleep(time.Second*time.Duration(ToExecuteMinus) + time.Second*time.Duration(ToExecutePlus) + time.Second*time.Duration(ToExecuteMultiply) + time.Second*time.Duration(ToExecuteDivide))
	return eval(expr)
}

func sendToOrchestraByGet(res float64) { // функция, необходимая для работы solve, отправляет решённое выражение
	addr := fmt.Sprintf("http://127.0.0.1:%s/receiveresult/?Result=%.3f&Id=%s&AgentPort=%s", ConnectedTo, Result, Id, AgentPort)
	fmt.Println(addr)
	//addr := "http://127.0.0.1:8080/receiveresult/?Result=15.5&Id=1488"
	_, _ = http.Get(addr)
}

func Connect(w http.ResponseWriter, r *http.Request) { // /connect/ Даёт порт, на котором хостится оркестр
	ConnectedTo = r.URL.Query().Get("HostPort")
	fmt.Println(ConnectedTo)

}

func Solve(w http.ResponseWriter, r *http.Request) { // /solve/ получает выражение и онправляет его оркустру
	Expression = r.URL.Query().Get("Expression")
	Id = r.URL.Query().Get("Id")
	ExecutionTimings := strings.Split(r.URL.Query().Get("ExecutionTimings"), "!")
	go func() {
		Result = evalWithDelay(Expression, ExecutionTimings)
		sendToOrchestraByGet(Result)
		fmt.Println("finished solving")
	}()
	fmt.Fprintln(w, "Started Solving")

}

func HandleHeratbeat(w http.ResponseWriter, r *http.Request) { // /heartbeat/ оркестр отправляет сюда хартбиты
	ConnectedTo = r.URL.Query().Get("HostPort")
	fmt.Println("hb received", ConnectedTo)
}

func main() {
	AgentPort = os.Args[1]
	if AgentPort == "" {
		log.Fatal("PORT not set")
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "you shouldnt be here)")
	})
	http.HandleFunc("/connect/", Connect)
	http.HandleFunc("/solve/", Solve)
	http.HandleFunc("/heartbeat/", HandleHeratbeat)

	http.ListenAndServe(":"+AgentPort, nil)
}

// 127.0.0.1:8999/connect/?HostPort=8080
// 127.0.0.1:8000/solve/?Expression=(2%2B2*5-3)%2F2&Id=1&ExecutionTimings=1!2!3!4
// + == %2B
// / == %2F
