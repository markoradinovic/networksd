package service

import (
	log "github.com/Sirupsen/logrus"
	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
	"net"
	"strconv"
	"sync"
	"time"
)

var (
	_, private24BitBlock, _ = net.ParseCIDR("10.0.0.0/8")
	_, private20BitBlock, _ = net.ParseCIDR("172.16.0.0/12")
	_, private16BitBlock, _ = net.ParseCIDR("192.168.0.0/16")
	OvelayNet               = "overlay"
)

type NetworkAddressService interface {
	Docker(ctx context.Context) (types.Info, error)
	Info(ctx context.Context) ([]types.NetworkResource, error)
	CreateNetwork(ctx context.Context, driver string, name string) (types.NetworkCreateResponse, error)
}

type networkConf struct {
	scope     *net.IPNet
	subnet    int
	blacklist []*net.IPNet
}

type networkAddressService struct {
	bridge  networkConf
	overlay networkConf
	mtx     sync.RWMutex
	cli     *client.Client
}

func NewNetworkAddressService(conf Conf) NetworkAddressService {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	return &networkAddressService{bridge: readConf(conf.Bridge), overlay: readConf(conf.Overlay), cli: cli}
}

func readConf(conf NetworkConf) (nc networkConf) {
	_, s, err := net.ParseCIDR(conf.NetworkScope)
	if err != nil {
		log.Panic(err)
	}
	var blacklist = make([]*net.IPNet, 0)
	for _, b := range conf.Blacklist {
		_, ipnet, err := net.ParseCIDR(b)
		if err != nil {
			log.Panic(err)
		}
		blacklist = append(blacklist, ipnet)
	}
	return networkConf{scope: s, subnet: conf.SubnetMask, blacklist: blacklist}
}

func (svc *networkAddressService) Docker(ctx context.Context) (types.Info, error) {
	return svc.cli.Info(ctx)
}

func (svc *networkAddressService) Info(ctx context.Context) ([]types.NetworkResource, error) {
	return listNetworks(ctx, svc.cli)
}

func (svc *networkAddressService) CreateNetwork(ctx context.Context, driver string, name string) (types.NetworkCreateResponse, error) {
	defer timeTrack(time.Now(), "CreateNetwork")

	svc.mtx.Lock()
	defer svc.mtx.Unlock()

	conf := svc.bridge
	if OvelayNet == driver {
		conf = svc.overlay
		info, err := svc.cli.Info(ctx)
		if err != nil {
			log.Error(err)
			return types.NetworkCreateResponse{}, err
		}
		//i, err := json.Marshal(info)
		//fmt.Println(string(i))
		if info.Swarm.LocalNodeState == "inactive" {
			return types.NetworkCreateResponse{}, ErrNotInSwarmMode
		}
	}

	//get all Docker Networks on Host
	dockerNetworks, err := listNetworks(ctx, svc.cli)
	if err != nil {
		return types.NetworkCreateResponse{}, err
	}

	//collect private networks
	existingNetworks := dockerNetworksToIPNets(dockerNetworks)

	//add blacklisted networks
	existingNetworks = append(existingNetworks, conf.blacklist...)

	debugExistingNetworks(existingNetworks)

	newNet := firstNetwork(conf)
	for {
		taken := isItTaken(newNet, existingNetworks)
		subnet := conf.scope.Contains(newNet.IP)
		private := isPrivateNet(newNet.IP)

		log.Debug("Testing network: [", newNet.String(), "]")
		log.Debug("Is it taken: [", newNet.String(), "] - ", taken)
		log.Debug("Is it valid subnet: [", newNet.String(), "] - ", subnet)
		log.Debug("Is it private network: [", newNet.String(), "] - ", private)

		if !taken && subnet && private {
			log.Info("Network to be created: [", newNet.String(), "]")
			return createNetwork(ctx, svc.cli, driver, name, newNet)
		}
		next, max := cidr.NextSubnet(newNet, conf.subnet)
		newNet = &net.IPNet{IP: next.IP, Mask: next.Mask}
		if max {
			return types.NetworkCreateResponse{}, ErrNoMoreNetworks
		}
	}
}

func listNetworks(ctx context.Context, cli *client.Client) ([]types.NetworkResource, error) {
	defer timeTrack(time.Now(), "listNetworks")
	return cli.NetworkList(ctx, types.NetworkListOptions{})
}

func dockerNetworksToIPNets(dockerNetworks []types.NetworkResource) []*net.IPNet {
	var ipNets = make([]*net.IPNet, 0)
	for _, dockerNetwork := range dockerNetworks {
		if dockerNetwork.Driver != "host" && dockerNetwork.Driver != "none" {
			subnet, found := getPrivateSubnet(dockerNetwork.IPAM.Config)
			if found {
				log.Debug("Found Docker Network: ", subnet)
				ipNets = append(ipNets, &subnet)
			}
		}
	}
	return ipNets
}

func createNetwork(ctx context.Context, cli *client.Client, driver string, name string, subnet *net.IPNet) (types.NetworkCreateResponse, error) {

	defer timeTrack(time.Now(), "Docker create network")
	log.Info("Creating Docker network with name: ", name)

	ipamSubnet := subnet.String()
	first, last := cidr.AddressRange(subnet)
	gateway := cidr.Inc(first)
	log.Debug("IP range: [", first, "-", last, "], Gateway: [", gateway, "]")

	ipamConfig := []network.IPAMConfig{
		{
			Subnet:  ipamSubnet,
			Gateway: gateway.String(),
		},
	}

	ipam := &network.IPAM{
		Driver: "default",
		Config: ipamConfig,
	}
	return cli.NetworkCreate(ctx, name, types.NetworkCreate{
		CheckDuplicate: true,
		EnableIPv6:     false,
		Driver:         driver,
		Attachable:     true,
		IPAM:           ipam,
	})
}

func isItTaken(network *net.IPNet, nets []*net.IPNet) bool {
	//defer timeTrack(time.Now(), "isItTaken")
	for _, n := range nets {
		if n.Contains(network.IP) {
			return true
		}
	}
	return false
}

func getPrivateSubnet(configs []network.IPAMConfig) (net.IPNet, bool) {
	//defer timeTrack(time.Now(), "getPrivateSubnet")
	var subnet net.IPNet
	for _, c := range configs {
		if c.Subnet != "" {
			_, s, err := net.ParseCIDR(c.Subnet)
			if err != nil {
				return *s, false
			}
			subnet = *s
			break
		}
	}
	if isPrivateNet(subnet.IP) {
		return subnet, true
	}
	return net.IPNet{}, false
}

func isPrivateNet(ip net.IP) bool {
	//defer timeTrack(time.Now(), "isPrivateNet")
	private := false
	private = private24BitBlock.Contains(ip) || private20BitBlock.Contains(ip) || private16BitBlock.Contains(ip)
	return private
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Debugf("%s took %s", name, elapsed)
}

func debugExistingNetworks(nets []*net.IPNet) {
	log.Debug("Existing networks: ")
	for _, net := range nets {
		log.Debug("[", net.String(), "]")
	}
}

func firstNetwork(conf networkConf) *net.IPNet {
	_, net, _ := net.ParseCIDR(conf.scope.IP.String() + "/" + strconv.Itoa(conf.subnet))
	return net
}
