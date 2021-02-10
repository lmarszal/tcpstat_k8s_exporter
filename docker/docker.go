package docker

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type PodSandbox struct {
	PodName   string
	Namespace string
	Pid       int
}

type ClientWithCache struct {
	isPodSandbox map[string]bool
	podSandboxes map[string]PodSandbox
	*client.Client
}

func New() (ClientWithCache, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return ClientWithCache{}, err
	}

	return ClientWithCache{
		isPodSandbox: map[string]bool{},
		podSandboxes: map[string]PodSandbox{},
		Client: cli,
	}, nil
}

func (cli *ClientWithCache) GetPodSandbox(ctx context.Context, containerId string) (sandbox PodSandbox, isSandbox bool, err error) {
	if isSandbox, found := cli.isPodSandbox[containerId]; found {
		if isSandbox {
			return cli.podSandboxes[containerId], true, nil
		} else {
			return PodSandbox{}, false, nil
		}
	}

	sandbox, isSandbox, err = cli.getPodSandbox(ctx, containerId)
	if err != nil {
		return PodSandbox{}, false, err
	}
	cli.isPodSandbox[containerId] = isSandbox
	if isSandbox {
		cli.podSandboxes[containerId] = sandbox
	}

	return sandbox, isSandbox, nil
}

func (cli *ClientWithCache) getPodSandbox(ctx context.Context, id string) (sandbox PodSandbox, isSandbox bool, err error) {
	info, err := cli.ContainerInspect(ctx, id)
	if err != nil {
		return PodSandbox{}, false, err
	}

	var found bool
	var dockerType string
	if dockerType, found = info.Config.Labels["io.kubernetes.docker.type"]; !found {
		return PodSandbox{}, false, nil
	}
	if dockerType != "podsandbox" {
		return PodSandbox{}, false, nil
	}

	var podName string
	if podName, found = info.Config.Labels["io.kubernetes.pod.name"]; !found {
		return PodSandbox{}, false, nil
	}
	var namespace string
	if namespace, found = info.Config.Labels["io.kubernetes.pod.namespace"]; !found {
		return PodSandbox{}, false, nil
	}

	sandbox = PodSandbox{
		PodName:   podName,
		Namespace: namespace,
		Pid:       info.State.Pid,
	}

	return sandbox, true, nil
}

func (cli *ClientWithCache) ListPodSandboxes(ctx context.Context, namespace string) ([]PodSandbox, error) {
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	sandboxes := make([]PodSandbox, 0)
	for _, c := range containers {
		sandbox, isSandbox, err := cli.GetPodSandbox(ctx, c.ID)
		if err != nil {
			return nil, err
		}
		if isSandbox && sandbox.Namespace == namespace {
			sandboxes = append(sandboxes, sandbox)
		}
	}

	return sandboxes, nil
}
