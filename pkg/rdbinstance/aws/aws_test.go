package aws

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/cloud-barista/mc-data-manager/pkg/rdbinstance"
)

func TestToDBInstances_MapsFields(t *testing.T) {
	in := []types.DBInstance{
		{
			DBInstanceIdentifier: awssdk.String("my-db"),
			Engine:               awssdk.String("mysql"),
			EngineVersion:        awssdk.String("8.0.35"),
			DBInstanceStatus:     awssdk.String("available"),
			DBInstanceClass:      awssdk.String("db.t3.micro"),
			Endpoint: &types.Endpoint{
				Address: awssdk.String("my-db.abc123.ap-northeast-2.rds.amazonaws.com"),
				Port:    awssdk.Int32(3306),
			},
		},
	}

	out := toDBInstances(in, "ap-northeast-2")

	if len(out) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(out))
	}
	got := out[0]
	if got.Provider != "aws" {
		t.Errorf("Provider = %q, want aws", got.Provider)
	}
	if got.InstanceID != "my-db" {
		t.Errorf("InstanceID = %q, want my-db", got.InstanceID)
	}
	if got.Name != "my-db" {
		t.Errorf("Name = %q, want my-db", got.Name)
	}
	if got.Engine != "mysql" {
		t.Errorf("Engine = %q, want mysql", got.Engine)
	}
	if got.EngineVersion != "8.0.35" {
		t.Errorf("EngineVersion = %q, want 8.0.35", got.EngineVersion)
	}
	if got.Status != "available" {
		t.Errorf("Status = %q, want available", got.Status)
	}
	if got.InstanceClass != "db.t3.micro" {
		t.Errorf("InstanceClass = %q, want db.t3.micro", got.InstanceClass)
	}
	if got.Endpoint != "my-db.abc123.ap-northeast-2.rds.amazonaws.com" {
		t.Errorf("Endpoint = %q", got.Endpoint)
	}
	if got.Port != 3306 {
		t.Errorf("Port = %d, want 3306", got.Port)
	}
	if got.Region != "ap-northeast-2" {
		t.Errorf("Region = %q, want ap-northeast-2", got.Region)
	}
}

func TestToDBInstances_NilEndpointDoesNotPanic(t *testing.T) {
	in := []types.DBInstance{
		{
			DBInstanceIdentifier: awssdk.String("pending-db"),
			DBInstanceStatus:     awssdk.String("creating"),
			Endpoint:             nil, // RDS omits endpoint until instance is ready
		},
	}

	out := toDBInstances(in, "ap-northeast-2")

	if len(out) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(out))
	}
	if out[0].Endpoint != "" {
		t.Errorf("Endpoint = %q, want empty", out[0].Endpoint)
	}
	if out[0].Port != 0 {
		t.Errorf("Port = %d, want 0", out[0].Port)
	}
}

func TestToDBInstance_MapsSingle(t *testing.T) {
	db := types.DBInstance{
		DBInstanceIdentifier: awssdk.String("new-db"),
		Engine:               awssdk.String("mariadb"),
		DBInstanceStatus:     awssdk.String("creating"),
		Endpoint:             nil, // creating instances have no endpoint yet
	}

	got := toDBInstance(db, "ap-northeast-2")

	if got.Provider != "aws" {
		t.Errorf("Provider = %q, want aws", got.Provider)
	}
	if got.InstanceID != "new-db" {
		t.Errorf("InstanceID = %q, want new-db", got.InstanceID)
	}
	if got.Engine != "mariadb" {
		t.Errorf("Engine = %q, want mariadb", got.Engine)
	}
	if got.Status != "creating" {
		t.Errorf("Status = %q, want creating", got.Status)
	}
	if got.Region != "ap-northeast-2" {
		t.Errorf("Region = %q", got.Region)
	}
	if got.Endpoint != "" || got.Port != 0 {
		t.Errorf("expected empty endpoint/port for creating instance, got %q/%d", got.Endpoint, got.Port)
	}
}

func TestBuildCreateInput_MapsSpec(t *testing.T) {
	spec := rdbinstance.CreateSpec{
		InstanceID:       "my-db",
		InstanceClass:    "db.t3.micro",
		Engine:           "mysql",
		EngineVersion:    "8.0.35",
		MasterUsername:   "admin",
		MasterPassword:   "secretpw",
		AllocatedStorage: 20,
	}

	in := buildCreateInput(spec)

	if in == nil {
		t.Fatal("expected non-nil input")
	}
	if awssdk.ToString(in.DBInstanceIdentifier) != "my-db" {
		t.Errorf("DBInstanceIdentifier = %q", awssdk.ToString(in.DBInstanceIdentifier))
	}
	if awssdk.ToString(in.DBInstanceClass) != "db.t3.micro" {
		t.Errorf("DBInstanceClass = %q", awssdk.ToString(in.DBInstanceClass))
	}
	if awssdk.ToString(in.Engine) != "mysql" {
		t.Errorf("Engine = %q", awssdk.ToString(in.Engine))
	}
	if awssdk.ToString(in.EngineVersion) != "8.0.35" {
		t.Errorf("EngineVersion = %q", awssdk.ToString(in.EngineVersion))
	}
	if awssdk.ToString(in.MasterUsername) != "admin" {
		t.Errorf("MasterUsername = %q", awssdk.ToString(in.MasterUsername))
	}
	if awssdk.ToString(in.MasterUserPassword) != "secretpw" {
		t.Errorf("MasterUserPassword = %q", awssdk.ToString(in.MasterUserPassword))
	}
	if awssdk.ToInt32(in.AllocatedStorage) != 20 {
		t.Errorf("AllocatedStorage = %d", awssdk.ToInt32(in.AllocatedStorage))
	}
	if !awssdk.ToBool(in.PubliclyAccessible) {
		t.Error("PubliclyAccessible = false, want true (fixed)")
	}
}

func TestToEngineVersions_TagsEngine(t *testing.T) {
	in := []types.DBEngineVersion{
		{EngineVersion: awssdk.String("8.0.35")},
		{EngineVersion: awssdk.String("8.0.36")},
	}

	out := toEngineVersions(in, "mysql")

	if len(out) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(out))
	}
	if out[0].Engine != "mysql" || out[0].EngineVersion != "8.0.35" {
		t.Errorf("out[0] = %+v, want {mysql 8.0.35}", out[0])
	}
	if out[1].Engine != "mysql" || out[1].EngineVersion != "8.0.36" {
		t.Errorf("out[1] = %+v, want {mysql 8.0.36}", out[1])
	}
}

func TestDistinctInstanceClasses_DedupAndSort(t *testing.T) {
	in := []types.OrderableDBInstanceOption{
		{DBInstanceClass: awssdk.String("db.t3.small")},
		{DBInstanceClass: awssdk.String("db.t3.micro")},
		{DBInstanceClass: awssdk.String("db.t3.micro")}, // duplicate across AZ/storage
		{DBInstanceClass: awssdk.String("db.t3.small")},
		{DBInstanceClass: awssdk.String("db.m5.large")},
	}

	out := distinctInstanceClasses(in)

	want := []string{"db.m5.large", "db.t3.micro", "db.t3.small"}
	if len(out) != len(want) {
		t.Fatalf("got %v, want %v", out, want)
	}
	for i := range want {
		if out[i] != want[i] {
			t.Errorf("out[%d] = %q, want %q (full: %v)", i, out[i], want[i], out)
		}
	}
}

func TestBuildDeleteInput_SkipsFinalSnapshot(t *testing.T) {
	in := buildDeleteInput("my-db")

	if in == nil {
		t.Fatal("expected non-nil input")
	}
	if awssdk.ToString(in.DBInstanceIdentifier) != "my-db" {
		t.Errorf("DBInstanceIdentifier = %q, want my-db", awssdk.ToString(in.DBInstanceIdentifier))
	}
	if !awssdk.ToBool(in.SkipFinalSnapshot) {
		t.Error("SkipFinalSnapshot = false, want true (fixed)")
	}
}
