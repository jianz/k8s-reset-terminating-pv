/*
Copyright (c) 2020 Jian Zhang
Licensed under MIT https://github.com/jianz/jianz.github.io/blob/master/LICENSE
*/

package cmd

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.etcd.io/etcd/clientv3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/protobuf"
)

var (
	etcdCA, etcdCert, etcdKey, etcdHost string
	etcdPort                            int

	k8sKeyPrefix string
	pvName       string

	cmd = &cobra.Command{
		Use:   "resetpv [flags] <persistent volume name>",
		Short: "Reset the Terminating PersistentVolume back to Bound status.",
		Long:  "Reset the Terminating PersistentVolume back to Bound status.\nPlease visit https://github.com/jianz/k8s-reset-terminating-pv for the detailed explanation.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("requires one persistent volume name argument")
			}
			pvName = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := resetPV()
			return err
		},
	}
)

// Execute reset the Terminating PersistentVolume to Bound status.
func Execute() {
	cmd.Flags().StringVar(&etcdCA, "etcd-ca", "ca.crt", "CA Certificate used by etcd")
	cmd.Flags().StringVar(&etcdCert, "etcd-cert", "etcd.crt", "Public key used by etcd")
	cmd.Flags().StringVar(&etcdKey, "etcd-key", "etcd.key", "Private key used by etcd")
	cmd.Flags().StringVar(&etcdHost, "etcd-host", "localhost", "The etcd domain name or IP")
	cmd.Flags().StringVar(&k8sKeyPrefix, "k8s-key-prefix", "registry", "The etcd key prefix for kubernetes resources.")
	cmd.Flags().IntVar(&etcdPort, "etcd-port", 2379, "The etcd port number")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func resetPV() error {
	etcdCli, err := etcdClient()
	if err != nil {
		return fmt.Errorf("cannot connect to etcd: %v", err)
	}
	defer etcdCli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return recoverPV(ctx, etcdCli)
}

func etcdClient() (*clientv3.Client, error) {
	ca, err := ioutil.ReadFile(etcdCA)
	if err != nil {
		return nil, err
	}

	keyPair, err := tls.LoadX509KeyPair(etcdCert, etcdKey)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(ca)

	return clientv3.New(clientv3.Config{
		Endpoints:   []string{fmt.Sprintf("%s:%d", etcdHost, etcdPort)},
		DialTimeout: 2 * time.Second,
		TLS: &tls.Config{
			RootCAs:      certPool,
			Certificates: []tls.Certificate{keyPair},
		},
	})
}

func recoverPV(ctx context.Context, client *clientv3.Client) error {

	gvk := schema.GroupVersionKind{Group: v1.GroupName, Version: "v1", Kind: "PersistentVolume"}
	pv := &v1.PersistentVolume{}

	runtimeScheme := runtime.NewScheme()
	runtimeScheme.AddKnownTypeWithName(gvk, pv)
	protoSerializer := protobuf.NewSerializer(runtimeScheme, runtimeScheme)

	// Get PV value from etcd which in protobuf format
	key := fmt.Sprintf("/%s/persistentvolumes/%s", k8sKeyPrefix, pvName)
	resp, err := client.Get(ctx, key)
	if err != nil {
		return err
	}

	if len(resp.Kvs) < 1 {
		return fmt.Errorf("cannot find persistent volume [%s] in etcd with key [%s]\nplease check the k8s-key-prefix and the persistent volume name are set correctly", pvName, key)
	}

	// Decode protobuf value to PV struct
	_, _, err = protoSerializer.Decode(resp.Kvs[0].Value, &gvk, pv)
	if err != nil {
		return err
	}

	// Set PV status from Terminating to Bound by removing value of DeletionTimestamp and DeletionGracePeriodSeconds
	if (*pv).ObjectMeta.DeletionTimestamp == nil {
		return fmt.Errorf("persistent volume [%s] is not in terminating status", pvName)
	}
	(*pv).ObjectMeta.DeletionTimestamp = nil
	(*pv).ObjectMeta.DeletionGracePeriodSeconds = nil

	// Encode fixed PV struct to protobuf value
	var fixedPV bytes.Buffer
	err = protoSerializer.Encode(pv, &fixedPV)
	if err != nil {
		return err
	}

	// Write the updated protobuf value back to etcd
	client.Put(ctx, key, fixedPV.String())
	return nil
}
