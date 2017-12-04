package main

import (
	"testing"
	"strings"
	"log"
)

func TestHandle(t *testing.T) {
  str:="filter=car_source_status= 0 and platform=1 and price>=100000 and city_id=12&from=0&query=&select=title& license_date& emission_standard& road_haul& price& image_url& effect_image_url& uni_code& city_id& clue_id&size=20&sort="
	keys:=strings.Replace(strings.Join([]string{str}, "&"), " ", "", -1)
	log.Println(keys)
}
