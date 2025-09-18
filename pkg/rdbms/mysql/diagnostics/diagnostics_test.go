package diagnostics_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/cloud-barista/mc-data-manager/pkg/rdbms/mysql/diagnostics"
)

func TestMain(t *testing.T) {
	sqlDB, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", "root", "N@mutech7^^7", "10.10.12.131", "3306"))
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	collector := diagnostics.NewCollector(sqlDB)
	// res2, err := collector.RunTimed(ctx, "mcmp", 30*time.Second)
	res2, err := collector.WithDiagnostic(ctx, func(ctx context.Context) error {
		// 여기서 실제 기능 실행: 예) 배치 작업, 마이그레이션, 대량 업데이트 등
		count()

		return nil
	})
	if err != nil { /* handle */
		panic(err)
	}

	fmt.Println(diagnostics.PrintBufferReport(res2.Buffer))
	fmt.Println(diagnostics.PrintWorkloadReport(res2.Work, res2.Elapsed))
	fmt.Println(diagnostics.PrintThreadReport(res2.Thread))
	fmt.Println(diagnostics.PrintLockReport(res2.Lock, res2.Elapsed))
	fmt.Println(diagnostics.PrintIOReport(res2.IO, res2.Elapsed))
	fmt.Println(res2.Report.String())
}

func count() {
	for i := 1; i <= 20; i++ {
		fmt.Printf("%d \n", i)
		time.Sleep(time.Second * 1)
	}
}
