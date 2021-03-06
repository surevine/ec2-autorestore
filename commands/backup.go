package commands

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
	"log"
)

func BackupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "backup [tagValue] [backupID]",
		Short: "Creates a snapshot of instances",
		Long:  `Creates a snapshot of instances that have the tag backup with a value of [tagValue] and gives it an ID of [backupID]`,
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			createSnapshots(args[0], args[1])
		},
	}
}

func createSnapshots(tagValue string, backupId string) {
	filters := []*ec2.Filter{new(ec2.Filter).SetName("tag:backup").SetValues([]*string{aws.String(tagValue)})}
	input := new(ec2.DescribeInstancesInput).SetFilters(filters)
	instancesMap := make(map[string]*ec2.Instance)

	for ok := true; ok; {
		out, err := ec2c.DescribeInstances(input)
		if err != nil {
			log.Fatal("Can't get instances", err)
		}
		for _, r := range out.Reservations {
			for _, i := range r.Instances {
				log.Printf("Found instance \"%v\"", *i.InstanceId)
				instancesMap[*i.InstanceId] = i
			}
		}
		if ok = out.NextToken != nil; ok {
			input.SetNextToken(*out.NextToken)
		}
	}

	snapshots, err := getSnapshots(backupId)
	if err != nil {
		log.Fatalf("Error checking for existing backups: %v", err)
	}
	if len(snapshots) > 0 {
		fmt.Printf("There are existing backups with the id \"%v\":\n", backupId)
		err = checkAndDeleteSnaps(snapshots)
		if err != nil {
			log.Fatalf("Could not delete snapshots because %v", err)
		}
	}

	ids := shutdownInstances(instancesMap)

	for id, i := range instancesMap {
		for _, b := range i.BlockDeviceMappings {
			if *b.DeviceName == *i.RootDeviceName {
				log.Printf("Snapshotting %s of %s", id, *b.Ebs.VolumeId)
				tags := append(i.Tags, new(ec2.Tag).SetKey("autorestore-backupId").SetValue(backupId))
				tags = append(tags, new(ec2.Tag).SetKey("autorestore-instanceId").SetValue(id))
				csi := new(ec2.CreateSnapshotInput).
					SetVolumeId(*b.Ebs.VolumeId).
					SetTagSpecifications([]*ec2.TagSpecification{new(ec2.TagSpecification).
						SetResourceType("snapshot").
						SetTags(tags)})
				snap, err := ec2c.CreateSnapshot(csi)
				if err != nil {
					log.Fatal("Failed to create snapshot", err)
				}
				log.Printf("Created %s\n", *snap.SnapshotId)
			}
		}
	}

	restartInstances(ids)
}
