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
	"os"
	"strconv"
	"time"
)

func main() {
	var fileConfig = flag.String("f", "schedule_config.yml", "configuration file")

	flag.Parse()

	cfg, err := config.NewScheduleConfig(*fileConfig)
	if err != nil {
		log.Fatalf("could not read config: %v", err)
	}

	defer GracefulShutdown()

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

	recover()
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
				ProjectID:  report.ProjectID,
				DatasetID:  report.DatasetID,
				TableID:    report.Table,
				ServiceKey: report.GoogleServiceKey,
			},
			GsConfig: &pb.GsConfig{
				SpreadsheetID: report.SpreadsheetID,
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
			log.Printf("Статус отчета: %v ", callsReq.IsOK)
			log.Printf("Предупреждения: %v ", callsReq.Warnings)
		}
	}
}

func GracefulShutdown() {
	if err := recover(); err != nil {
		fmt.Println("Критическая ошибка:", err)
	}

	os.Exit(0)
}
