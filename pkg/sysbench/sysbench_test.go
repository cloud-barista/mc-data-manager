package sysbench_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cloud-barista/mc-data-manager/pkg/sysbench"
)

func TestMain(t *testing.T) {
	mysql_host := "10.10.12.131"
	mysql_port := "3306"
	mysql_user := "root"
	mysql_password := "N@mutech7^^7"
	mysql_database := "inventory"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	table_count := 1
	table_size := 1000
	threads_count := 8
	time := 10

	_, prepareErr := sysbench.RunSysbench(ctx,
		"mysql",
		false,
		"oltp_read_write",
		"--db-driver=mysql",
		"--mysql-host="+mysql_host,
		"--mysql-port="+mysql_port,
		"--mysql-user="+mysql_user,
		"--mysql-password="+mysql_password,
		"--mysql-db="+mysql_database,
		fmt.Sprintf("--tables=%d", table_count),
		fmt.Sprintf("--table-size=%d", table_size),
		fmt.Sprintf("--threads=%d", threads_count),
		fmt.Sprintf("--time=%d", time),
		"prepare",
	)
	if prepareErr != nil {
		fmt.Println("sysbench error:", prepareErr)
		// 그래도 res.Raw 안에 원문이 있으니 디버깅 가능
	}

	// 예: sysbench oltp_read_write --threads=8 --time=10 --report-interval=1 run
	res, err := sysbench.RunSysbench(ctx,
		"mysql",
		true,
		"oltp_read_write",
		"--db-driver=mysql",
		"--mysql-host="+mysql_host,
		"--mysql-port="+mysql_port,
		"--mysql-user="+mysql_user,
		"--mysql-password="+mysql_password,
		"--mysql-db="+mysql_database,
		fmt.Sprintf("--tables=%d", table_count),
		fmt.Sprintf("--table-size=%d", table_size),
		fmt.Sprintf("--threads=%d", threads_count),
		fmt.Sprintf("--time=%d", time),
		"run",
	)
	if err != nil {
		fmt.Println("sysbench error:", err)
		// 그래도 res.Raw 안에 원문이 있으니 디버깅 가능
	}

	_, cleanErr := sysbench.RunSysbench(ctx,
		"mysql",
		false,
		"oltp_read_write",
		"--db-driver=mysql",
		"--mysql-host="+mysql_host,
		"--mysql-port="+mysql_port,
		"--mysql-user="+mysql_user,
		"--mysql-password="+mysql_password,
		"--mysql-db="+mysql_database,
		fmt.Sprintf("--tables=%d", table_count),
		fmt.Sprintf("--table-size=%d", table_size),
		fmt.Sprintf("--threads=%d", threads_count),
		fmt.Sprintf("--time=%d", time),
		"cleanup",
	)
	if cleanErr != nil {
		fmt.Println("sysbench error:", cleanErr)
		// 그래도 res.Raw 안에 원문이 있으니 디버깅 가능
	}

	fmt.Println(res.FormatSysbenchLike())
}
