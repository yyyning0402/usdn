package openflow

import (
	"context"
	"os/exec"
	"time"
)

//执行linux命令并获取标准输出和标准错误
func Cmd(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.CombinedOutput()
	result := string(out)
	return result, err
}

//执行linux命令获取标准输出和标准错误并设置超时时间
func CmdTime(command string, ts time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*ts)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	out, err := cmd.CombinedOutput()
	result := string(out)
	return result, err
}

// func main () {
// 	re,err := Cmd("echo hello word")
// 	fmt.Println(re,err)
// 	re,err = CmdTime("echo hello;sleep 6",5)
// 	fmt.Println("123",re,"sss",err)
// }
