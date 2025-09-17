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
	res2, err := collector.WithDiagnostic(ctx, "inventory", func(ctx context.Context) error {
		// 여기서 실제 기능 실행: 예) 배치 작업, 마이그레이션, 대량 업데이트 등
		count()

		return nil
	})
	if err != nil { /* handle */
		panic(err)
	}

	diagnostics.PrintBufferReport(res2.Buffer)
	diagnostics.PrintLockReport(res2.Lock, res2.Elapsed)
	diagnostics.PrintIOReport(res2.IO, res2.Elapsed)
	diagnostics.PrintWorkloadReport(res2.Work)
	diagnostics.PrintThreadReport(res2.Thread)
}

func count() {
	for i := 1; i <= 10; i++ {
		fmt.Printf("%d \n", i)
		time.Sleep(time.Second * 1)
	}
}
