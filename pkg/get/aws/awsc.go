package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cloud-barista/mc-data-manager/internal/auth"
	"github.com/cloud-barista/mc-data-manager/models"
)

// AWSClient wraps aws.Config and provides additional methods
type AWSClient struct {
	Config aws.Config
}

func newAWSConfig(params models.AWSCredentials) (*AWSClient, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(params.AccessKey, params.SecretKey, "")),
		config.WithRetryMaxAttempts(5),
	)

	if err != nil {
		return nil, err
	}

	return &AWSClient{Config: cfg}, nil
}

// ListS3Buckets lists all S3 buckets in the specified region
func (client *AWSClient) ListS3Buckets() ([]string, error) {
	svc := s3.NewFromConfig(client.Config)
	resp, err := svc.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list buckets, %v", err)
	}

	buckets := make([]string, len(resp.Buckets))
	for i, bucket := range resp.Buckets {
		buckets[i] = *bucket.Name
	}

	return buckets, nil
}

// ListDynamoDBTables lists all DynamoDB tables in the specified region
func (client *AWSClient) ListDynamoDBTables() ([]string, error) {
	svc := dynamodb.NewFromConfig(client.Config)
	resp, err := svc.ListTables(context.TODO(), &dynamodb.ListTablesInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list tables, %v", err)
	}

	return resp.TableNames, nil
}

// ListEC2Instances lists all EC2 instances in the specified region
func (client *AWSClient) ListEC2Instances() ([]string, error) {
	svc := ec2.NewFromConfig(client.Config)
	resp, err := svc.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list instances, %v", err)
	}

	var instances []string
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			instances = append(instances, *instance.InstanceId)
		}
	}

	return instances, nil
}

// ListEC2InstanceTypes lists all EC2 instance types (flavors) in the specified region
func (client *AWSClient) ListEC2InstanceTypes() ([]string, error) {
	svc := ec2.NewFromConfig(client.Config)
	resp, err := svc.DescribeInstanceTypes(context.TODO(), &ec2.DescribeInstanceTypesInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list instance types, %v", err)
	}

	var instanceTypes []string
	for _, instanceType := range resp.InstanceTypes {
		instanceTypes = append(instanceTypes, string(instanceType.InstanceType))
	}

	return instanceTypes, nil
}

// ListEC2InstancesByRegion lists all EC2 instances in the specified Region
func (client *AWSClient) ListEC2InstancesByRegion(region string) ([]string, error) {
	client.Config.Region = region
	svc := ec2.NewFromConfig(client.Config)
	resp, err := svc.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list instances by region, %v", err)
	}

	var instances []string
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			instances = append(instances, *instance.InstanceId)
		}
	}

	return instances, nil
}

// ListEC2InstancesByVPC lists all EC2 instances in the specified VPC
func (client *AWSClient) ListEC2InstancesByVPC(vpcID string) ([]string, error) {
	svc := ec2.NewFromConfig(client.Config)
	resp, err := svc.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcID},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to list instances by VPC, %v", err)
	}

	var instances []string
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			instances = append(instances, *instance.InstanceId)
		}
	}

	return instances, nil
}

// ListSubnetsByVPC lists all subnets in the specified VPC
func (client *AWSClient) ListSubnetsByVPC(vpcID string) ([]string, error) {
	svc := ec2.NewFromConfig(client.Config)
	resp, err := svc.DescribeSubnets(context.TODO(), &ec2.DescribeSubnetsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcID},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to list subnets by VPC, %v", err)
	}

	var subnets []string
	for _, subnet := range resp.Subnets {
		subnets = append(subnets, *subnet.SubnetId)
	}

	return subnets, nil
}

// ListRegions lists all available regions
func (client *AWSClient) ListRegions() ([]string, error) {
	tempCfg := client.Config
	tempCfg.Region = "us-west-2"
	svc := ec2.NewFromConfig(tempCfg)
	resp, err := svc.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list regions, %v", err)
	}

	var regions []string
	for _, region := range resp.Regions {
		regions = append(regions, *region.RegionName)
	}

	return regions, nil
}

func main() {
	fmt.Println("Starting main function")

	profileName := "default"
	provider := "aws"
	defaultRegion := "ap-northeast-2"

	fmt.Println("Creating CredentialsManager")
	credentialsManager := auth.NewFileCredentialsManager()

	fmt.Println("Loading credentials")
	// Load credentials for the specified profile and provider
	creds, err := credentialsManager.LoadCredentialsByProfile(profileName, provider)
	if err != nil {
		fmt.Println("Error loading credentials:", err)
		return
	}

	fmt.Println("Casting credentials")
	awsCreds, ok := creds.(models.AWSCredentials)
	if !ok {
		fmt.Println(creds)
		fmt.Println("Invalid credentials type")
		return
	}

	fmt.Println("Creating AWS config")
	client, err := newAWSConfig(awsCreds)
	if err != nil {
		fmt.Println("Error creating AWS config:", err)
		return
	}

	regions, err := client.ListRegions()
	if err != nil {
		fmt.Println("Error listing regions:", err)
		return
	}
	fmt.Println("Regions:", regions)

	client.Config.Region = defaultRegion
	fmt.Println("Listing AWS resources")

	// List AWS resources
	listResources(client)

	fmt.Println("Finished main function")
}

// listResources lists various AWS resources
func listResources(client *AWSClient) {
	totalSteps := 6
	currentStep := 0

	fmt.Println("Listing S3 Buckets")
	// List S3 Buckets
	buckets, err := client.ListS3Buckets()
	if err != nil {
		fmt.Println("Error listing S3 buckets:", err)
		return
	}
	fmt.Println("S3 Buckets:", buckets)
	currentStep++
	progressBar(currentStep, totalSteps)

	fmt.Println("Listing DynamoDB Tables")
	// List DynamoDB Tables
	tables, err := client.ListDynamoDBTables()
	if err != nil {
		fmt.Println("Error listing DynamoDB tables:", err)
		return
	}
	fmt.Println("DynamoDB Tables:", tables)
	currentStep++
	progressBar(currentStep, totalSteps)

	fmt.Println("Listing EC2 Instances")
	// List EC2 Instances
	instances, err := client.ListEC2Instances()
	if err != nil {
		fmt.Println("Error listing EC2 instances:", err)
		return
	}
	fmt.Println("EC2 Instances:", instances)
	currentStep++
	progressBar(currentStep, totalSteps)

	fmt.Println("Listing EC2 Instance Types")
	// List EC2 Instance Types
	instanceTypes, err := client.ListEC2InstanceTypes()
	if err != nil {
		fmt.Println("Error listing EC2 instance types:", err)
		return
	}
	fmt.Println("EC2 Instance Types:", instanceTypes)
	currentStep++
	progressBar(currentStep, totalSteps)

	fmt.Println("Listing EC2 Instances by VPC")
	// List EC2 Instances by VPC
	vpcID := "your-vpc-id"
	instancesByVPC, err := client.ListEC2InstancesByVPC(vpcID)
	if err != nil {
		fmt.Println("Error listing EC2 instances by VPC:", err)
		return
	}
	fmt.Println("EC2 Instances by VPC:", instancesByVPC)
	currentStep++
	progressBar(currentStep, totalSteps)

	fmt.Println("Listing Subnets by VPC")
	// List Subnets by VPC
	subnetsByVPC, err := client.ListSubnetsByVPC(vpcID)
	if err != nil {
		fmt.Println("Error listing subnets by VPC:", err)
		return
	}
	fmt.Println("Subnets by VPC:", subnetsByVPC)
	currentStep++
	progressBar(currentStep, totalSteps)
}

// Progress bar function
func progressBar(current, total int) {
	percent := float64(current) / float64(total) * 100
	bar := "["
	for i := 0; i < 50; i++ {
		if i < int(percent/2) {
			bar += "="
		} else {
			bar += " "
		}
	}
	bar += "]"
	fmt.Printf("\r%s %3.0f%%", bar, percent)
	if current == total {
		fmt.Println()
	}
}
