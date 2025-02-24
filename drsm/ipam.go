// SPDX-FileCopyrightText: 2022 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
package drsm

import (
	"context"
	"fmt"
	ipam "github.com/thakurajayL/go-ipam"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

// TODO : should have ability to create new instances of ipam
func (d *Drsm) initIpam(opt *Options) {
	if opt != nil {
		log.Println("ipmodule ", opt)
	}
	dbOptions := &options.ClientOptions{}
	dbOptions = dbOptions.ApplyURI(d.db.Url)
	dbConfig := ipam.MongoConfig{DatabaseName: d.db.Name, CollectionName: "ipaddress", MongoClientOptions: dbOptions}
	mo, err := ipam.NewMongo(context.TODO(), dbConfig)
	if err != nil {
		log.Println("ipmodule error. NewMongo error  ", err)
	}
	ipModule := ipam.NewWithStorage(mo)
	log.Println("ipmodule ", ipModule)
	d.ipModule = ipModule
	d.prefix = make(map[string]*ipam.Prefix)

	for k, v := range opt.IpPool {
		prefix, err := ipModule.NewPrefix(context.TODO(), v)
		if err != nil {
			panic(err)
		}
		d.prefix[k] = prefix
	}
	log.Println("ip module prefix ", d.prefix)
}

func (d *Drsm) initIpPool(name string, prefix string) error {
	p, err := d.ipModule.NewPrefix(context.TODO(), prefix)
	if err != nil {
		return err
	}
	d.prefix[name] = p
	return nil
}

func (d *Drsm) deleteIpPool(name string) error {
	p, found := d.prefix[name]
	if found == false {
		err := fmt.Errorf("Failed to find pool %s", name)
		return err
	}
	_, err := d.ipModule.DeletePrefix(context.TODO(), p.Cidr)
	return err
}

func (d *Drsm) acquireIp(name string) (string, error) {
	prefix, found := d.prefix[name]
	if found != true {
		err := fmt.Errorf("IP Pool %v not found ", name)
		return "", err
	}

	ip, err := d.ipModule.AcquireIP(context.TODO(), prefix.Cidr)
	if err != nil {
		err := fmt.Errorf("No address")
		return "", err
	}
	log.Println("Acquired IP ", ip.IP)
	return ip.IP.String(), nil
}

func (d *Drsm) releaseIp(name, ip string) error {
	prefix, found := d.prefix[name]
	if found != true {
		err := fmt.Errorf("IP Pool %v not found ", name)
		return err
	}

	err := d.ipModule.ReleaseIPFromPrefix(context.TODO(), prefix.Cidr, ip)
	if err != nil {
		log.Println("Release IP failed - ", ip)
		err := fmt.Errorf("No address")
		return err
	}
	log.Println("Release IP successful ", ip)
	return nil
}
