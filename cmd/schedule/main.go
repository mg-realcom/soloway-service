package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/go-co-op/gocron"
	pb "github.com/mg-realcom/go-genproto/service/soloway.v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"soloway/internal/config"
)

func main() {
	fileConfig := flag.String("f", "schedule_config.yml", "configuration file")

	flag.Parse()

	cfg, err := config.NewScheduleConfig(*fileConfig)
	if err != nil {
		log.Fatalf("could not read config: %v", err)
	}

	fmt.Printf("Количество отчетов: %d\n", len(cfg.Reports))
	fmt.Printf("Запуск ежедневно в: %s\n", cfg.Time)

	s := gocron.NewScheduler(time.UTC)
	s.WaitForScheduleAll()

	location, err := time.LoadLocation("Local")
	if err != nil {
		log.Fatalf("could not load location: %v", err)
	}

	s.ChangeLocation(location)

	_, err = s.Every(1).Day().At(cfg.Time).Do(scheduleRun, *cfg)
	if err != nil {
		log.Fatalf("Ошибка планировщика %v", err)
	}

	s.StartBlocking()
}

func scheduleRun(cfg config.ScheduleConfig) {
	conn, err := grpc.Dial(cfg.Destination, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
			CsConfig: &pb.CsConfig{
				BucketName: "",
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
			log.Printf("Предупреждения: %v ", callsReq.Warnings)
		}
	}
}
