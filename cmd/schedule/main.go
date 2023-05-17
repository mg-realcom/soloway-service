package main

import (
	"Soloway/internal/config"
	"Soloway/pb"
	"context"
	"flag"
	"fmt"
	"github.com/go-co-op/gocron"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"strconv"
	"time"
)

const version = "1.1.0"

func main() {
	var fileConfig = flag.String("f", "schedule_config.yml", "configuration file")

	flag.Parse()

	cfg, err := config.NewScheduleConfig(*fileConfig)
	if err != nil {
		log.Fatalf("could not read config: %v", err)
	}

	fmt.Printf("version: %s\n", version)

	fmt.Printf("Количество отчетов: %d\n", len(cfg.Reports))
	fmt.Printf("Запуск ежедневно в: %s\n", cfg.Time)

	s := gocron.NewScheduler(time.UTC)
	s.WaitForScheduleAll()

	location, err := time.LoadLocation("Local")

	s.ChangeLocation(location)

	_, err = s.Every(1).Day().At(cfg.Time).Do(scheduleRun, *cfg)
	if err != nil {
		log.Fatalf("Ошибка планировщика %v", err)
	}

	s.StartBlocking()
}

func scheduleRun(cfg config.ScheduleConfig) {
	addr := net.JoinHostPort(cfg.GRPC.IP, strconv.Itoa(cfg.GRPC.Port))

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
	}(conn)

	c := pb.NewSolowayServiceClient(conn)

	ctx := context.Background()

	for _, report := range cfg.Reports {
		log.Printf("Сбор отчета для: %s\n", report.SpreadsheetID)

		dateTill := time.Now()
		dateFrom := dateTill.AddDate(0, 0, -report.Days)

		callsReq, err := c.PushPlacementStatByDayToBQ(ctx, &pb.PushPlacementStatByDayToBQRequest{
			BqConfig: &pb.BqConfig{
				ProjectId:  report.ProjectID,
				DatasetId:  report.DatasetID,
				TableId:    report.Table,
				ServiceKey: report.GoogleServiceKey,
			},
			GsConfig: &pb.GsConfig{
				SpreadsheetId: report.SpreadsheetID,
				ServiceKey:    report.GoogleServiceKey,
			},
			Period: &pb.Period{
				DateFrom: dateFrom.Format("2006-01-02"),
				DateTill: dateTill.Format("2006-01-02"),
			},
		})

		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Статус отчета: %v ", callsReq.IsOk)
			log.Printf("Предупреждения: %v ", callsReq.Warnings)
		}
	}
}
