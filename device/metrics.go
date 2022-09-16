package device

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.zx2c4.com/wireguard/conn"
)

var (
	peerBytesReceived = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "wireguard_peer_received_bytes_total",
		Help: "Total number of received bytes",
	}, []string{"peer", "us", "them"})
	peerPacketsReceived = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "wireguard_peer_received_packets_total",
		Help: "Total number of received packets",
	}, []string{"peer", "us", "them"})
	peerBytesDropped = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "wireguard_peer_deduplicated_bytes_total",
		Help: "Total number of received bytes dropped by deduplication",
	}, []string{"peer", "us", "them"})
	peerPacketsDropped = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "wireguard_peer_deduplicated_packets_total",
		Help: "Total number of received packets dropped by deduplication",
	}, []string{"peer", "us", "them"})
	peerBytesSent = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "wireguard_peer_sent_bytes_total",
		Help: "Total number of received bytes",
	}, []string{"peer", "us", "them"})
	peerPacketsSent = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "wireguard_peer_sent_packets_total",
		Help: "Total number of received packets",
	}, []string{"peer", "us", "them"})
)

func metricLabels(peer *Peer, endpoint conn.Endpoint) prometheus.Labels {
	return prometheus.Labels{
		"peer": peer.String(),
		"us":   endpoint.SrcToString(),
		"them": endpoint.DstToString(),
	}
}
