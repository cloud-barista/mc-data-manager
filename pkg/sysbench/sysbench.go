package sysbench

import (
	"context"
	"fmt"
	"os/exec"
)

type SysbenchParsed struct {
	TargetType  string               `json:"targetType"`
	RdbmsResult *MySQLSysbenchParsed `json:"rdbmsResult,omitempty"`
}

/***** 실행 + 파싱 *****/

// RunSysbench 실행: ctx로 타임아웃/취소 제어, sysbench args를 그대로 전달
func RunSysbench(ctx context.Context, targetType string, doParse bool, args ...string) (SysbenchParsed, error) {
	cmd := exec.CommandContext(ctx, "sysbench", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// sysbench 실행 실패(커맨드 자체 실패)와 벤치 실패를 구분하기 어렵기 때문에
		// 원문 출력과 함께 에러 반환
		return SysbenchParsed{TargetType: targetType}, fmt.Errorf("sysbench failed: %w; output: %s", err, string(out))
	}

	if !doParse {
		return SysbenchParsed{TargetType: targetType}, nil
	}
	return ParseSysbenchOutput(targetType, out)
}

func ParseSysbenchOutput(targetType string, out []byte) (SysbenchParsed, error) {
	var result SysbenchParsed = SysbenchParsed{}
	switch targetType {
	case "mysql":
		rdbmsResult, mErr := ParseMySQLSysbenchOutput(out)
		if mErr != nil {
			return result, mErr
		}
		result.TargetType = "mysql"
		result.RdbmsResult = &rdbmsResult
		return result, nil
	}

	return result, nil
}

func (s SysbenchParsed) FormatSysbenchLike() string {
	var result string = ""
	switch s.TargetType {
	case "mysql":
		result = s.RdbmsResult.FormatMySQLSysbenchLike()
	}

	return result
}
