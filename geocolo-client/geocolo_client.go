/*-
 * Copyright (c) 2012 Caoimhe Chaos <caoimhechaos@protonmail.com>,
 *                    Ancient Solutions. All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 *
 * 1. Redistributions  of source code must retain  the above copyright
 *    notice, this list of conditions and the following disclaimer.
 * 2. Redistributions  in   binary  form  must   reproduce  the  above
 *    copyright  notice, this  list  of conditions  and the  following
 *    disclaimer in the  documentation and/or other materials provided
 *    with the distribution.
 *
 * THIS  SOFTWARE IS  PROVIDED BY  ANCIENT SOLUTIONS  AND CONTRIBUTORS
 * ``AS IS'' AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
 * LIMITED TO,  THE IMPLIED WARRANTIES OF  MERCHANTABILITY AND FITNESS
 * FOR A  PARTICULAR PURPOSE  ARE DISCLAIMED.  IN  NO EVENT  SHALL THE
 * FOUNDATION  OR CONTRIBUTORS  BE  LIABLE FOR  ANY DIRECT,  INDIRECT,
 * INCIDENTAL,   SPECIAL,    EXEMPLARY,   OR   CONSEQUENTIAL   DAMAGES
 * (INCLUDING, BUT NOT LIMITED  TO, PROCUREMENT OF SUBSTITUTE GOODS OR
 * SERVICES; LOSS OF USE,  DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
 * HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT,
 * STRICT  LIABILITY,  OR  TORT  (INCLUDING NEGLIGENCE  OR  OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED
 * OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"strings"

	"ancient-solutions.com/net/geocolo"
	"ancient-solutions.com/net/urlconnection"
)

func main() {
	var endpoint, uri, buri, origin, candidates string
	var maxdistance float64
	var client *rpc.Client
	var mode string
	var conn net.Conn
	var detailed bool
	var err error

	flag.StringVar(&endpoint, "endpoint", "",
		"The service URL to connect to")
	flag.StringVar(&uri, "doozer-uri", os.Getenv("DOOZER_URI"),
		"Doozer URI to connect to")
	flag.StringVar(&buri, "doozer-boot-uri", os.Getenv("DOOZER_BOOT_URI"),
		"Doozer Boot URI to find named clusters")
	flag.StringVar(&origin, "origin", "",
		"Country which we're looking for close countries for")
	flag.StringVar(&candidates, "candidates", "",
		"Comma separated list of countries to consider")
	flag.StringVar(&mode, "mode", "country",
		"Method to contact (country or ip)")
	flag.Float64Var(&maxdistance, "max-distance", 0,
		"Maximum distance from the closest IP to consider")
	flag.BoolVar(&detailed, "detailed", false,
		"Whether to give a detailed response")
	flag.Parse()

	if uri != "" {
		if err = urlconnection.SetupDoozer(buri, uri); err != nil {
			log.Fatal("Error initializing Doozer connection to ",
				uri, ": ", err.Error())
		}
	}

	conn, err = urlconnection.Connect(endpoint)
	if err != nil {
		log.Fatal("Error connecting to ", endpoint, ": ", err.Error())
	}
	client = rpc.NewClient(conn)

	if mode == "country" {
		var req geocolo.GeoProximityRequest
		var res geocolo.GeoProximityResponse

		if len(candidates) > 0 {
			req.Candidates = strings.Split(candidates, ",")
		}

		req.Origin = &origin
		req.DetailedResponse = &detailed

		err = client.Call("GeoProximityService.GetProximity", req,
			&res)
		if err != nil {
			log.Fatal("Error sending proximity request: ",
				err.Error())
		}

		if res.Closest == nil {
			log.Fatal("Failed to fetch closest country")
		} else {
			fmt.Printf("Closest country: %s\n", *res.Closest)
		}

		for _, detail := range res.FullMap {
			if detail == nil {
				log.Print("Error: detail is nil?")
			} else if detail.Country == nil {
				log.Print("Error: country is nil?")
				if detail.Distance != nil {
					log.Printf("(distance was %f)",
						*detail.Distance)
				}
			} else if detail.Distance == nil {
				log.Print("Error: distance is nil?")
				if detail.Country != nil {
					log.Printf("(country was %s)", *detail.Country)
				}
			} else {
				fmt.Printf("Country %s: distance %f\n", *detail.Country,
					*detail.Distance)
			}
		}
	} else if mode == "ip" {
		var req geocolo.GeoProximityByIPRequest
		var res geocolo.GeoProximityByIPResponse

		req.Candidates = strings.Split(candidates, ",")
		req.DetailedResponse = &detailed
		req.Origin = &origin
		req.MaxDistance = &maxdistance

		err = client.Call("GeoProximityService.GetProximityByIP",
			req, &res)
		if err != nil {
			log.Fatal("Error sending proximity request: ",
				err.Error())
		}

		for _, addr := range res.Closest {
			fmt.Printf("Close IP: %s\n", addr)
		}

		for _, detail := range res.FullMap {
			fmt.Printf("IP: %s, distance: %f\n", *detail.Ip,
				*detail.Distance)
		}
	}
}