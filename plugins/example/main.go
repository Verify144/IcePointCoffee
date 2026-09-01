// 示例插件 - 监听 :8766 端口的 HTTP RPC 服务
// 编译方式：go build -o plugin main.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type RPCRequest struct {
	Method string         `json:"method"`
	Params map[string]any `json:"params"`
}

type RPCResponse struct {
	Result any    `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

func main() {
	http.HandleFunc("/rpc", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req RPCRequest
		if err := json.Unmarshal(body, &req); err != nil {
			writeResp(w, RPCResponse{Error: err.Error()})
			return
		}

		fmt.Printf("[plugin] method=%s params=%v\n", req.Method, req.Params)

		var result any
		switch req.Method {
		case "ping":
			result = "pong"
		case "echo":
			result = req.Params
		case "build":
			// 自定义建筑逻辑
			result = map[string]any{
				"status": "ok",
				"type":   req.Params["type"],
			}
		default:
			writeResp(w, RPCResponse{Error: "unknown method: " + req.Method})
			return
		}

		writeResp(w, RPCResponse{Result: result})
	})

	fmt.Println("示例插件启动: 127.0.0.1:8766")
	http.ListenAndServe("127.0.0.1:8766", nil)
}

func writeResp(w http.ResponseWriter, resp RPCResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
