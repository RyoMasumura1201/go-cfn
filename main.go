package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/ecs"
)
func main() {

	// dynamodbからデータ取得
	ddb := dynamodb.New(session.New(), aws.NewConfig().WithRegion("ap-northeast-1"))

	params := &dynamodb.ScanInput{
		TableName: aws.String("version_test"),

		AttributesToGet: []*string{
			aws.String("version"),
		},
		ConsistentRead: aws.Bool(true),

		ReturnConsumedCapacity: aws.String("NONE"),
	}

	resp, err := ddb.Scan(params)

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(*resp.Items[0]["version"].S)

	version1 := *resp.Items[0]["version"].S

	// yaml作成
	file, err := os.Create("ecs.yaml")

	if err != nil {
		fmt.Println(err)
	}
	template := cloudformation.NewTemplate()

	// クラスター作成
	clusterName := "mycluster"
	template.Resources["Cluster"] = &ecs.Cluster{
		ClusterName: &clusterName,
	}

	// タスク定義作成
	familyName := "tasktest"
	networkMode := "awsvpc"
	exectionRoleArn := cloudformation.ImportValue("aa")
	envKey := "ENV"
	envValue := "STG" 
	taskDefinitionResourceName := "TaskDefinition" + version1
	template.Resources[taskDefinitionResourceName] = &ecs.TaskDefinition{
		Family: &familyName,
		NetworkMode: &networkMode,
		ExecutionRoleArn: &exectionRoleArn,
		ContainerDefinitions: []ecs.TaskDefinition_ContainerDefinition{
			{
				Name: "task",
				Image: cloudformation.ImportValue("bb"),
				Environment: []ecs.TaskDefinition_KeyValuePair{
					{
					Name: &envKey,
					Value: &envValue,
					},
				},

			},
		},
		RequiresCompatibilities: []string{"FARGATE"},
	}

	// サービス作成
	serviceResourceName := "Service" + version1
	serviceName := "service" + version1
	clusterRef := cloudformation.Ref("Cluster")
	launchType := "FARGATE"
	taskDefinition := cloudformation.Ref("TaskDefinitionVersion1")
	desireCount := 1

	template.Resources[serviceResourceName] = &ecs.Service{
		ServiceName: &serviceName,
		Cluster: &clusterRef,
		LaunchType: &launchType,
		TaskDefinition: &taskDefinition,
		DesiredCount: &desireCount,
	}

	// 出力
	y, err := template.YAML()
	if err != nil {
		fmt.Println(err)
	}
	file.Write(y)
}