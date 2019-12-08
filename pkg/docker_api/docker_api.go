package docker_api

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
)

type DockerClient struct {
	client *client.Client
}

func NewDockerClient(host string) (*DockerClient, error) {
	cnt, err := client.NewClient(host, client.DefaultVersion, nil, nil)
	if err != nil {
		return nil, err
	}
	return &DockerClient{
		client: cnt,
	}, nil
}

func (d *DockerClient) CreateSwarmService(name string, image string, replicas uint64, publish map[uint32]uint32) error {
	err := d.RemoveSwarmService(name)
	if err != nil {
		return err
	}

	containerSpec := swarm.ContainerSpec{
		Image: image,
	}
	taskSpec := swarm.TaskSpec{
		ContainerSpec: containerSpec,
	}
	var pointConfigs []swarm.PortConfig
	for t, p := range publish {
		pointConfigs = append(pointConfigs, swarm.PortConfig{
			Protocol:      "tcp",
			TargetPort:    t,
			PublishedPort: p,
		})
	}
	serviceSpec := swarm.ServiceSpec{
		Annotations:  swarm.Annotations{Name: name},
		TaskTemplate: taskSpec,
		Mode:         swarm.ServiceMode{Replicated: &swarm.ReplicatedService{Replicas: &replicas}},
		EndpointSpec: &swarm.EndpointSpec{Ports: pointConfigs},
	}

	_, err = d.client.ServiceCreate(context.Background(), serviceSpec, types.ServiceCreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (d *DockerClient) RemoveSwarmService(name string) error {
	f := filters.NewArgs()
	f.Add("name", name)
	services, err := d.client.ServiceList(context.Background(), types.ServiceListOptions{Filters: f})
	if err != nil {
		return err
	}
	for _, s := range services {
		err = d.client.ServiceRemove(context.Background(), s.ID)
	}
	if err != nil {
		return err
	}
	return nil
}
