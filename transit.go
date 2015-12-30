package transitspark

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
	"net/http"
	"os"
	"sqbu-github.cisco.com/jgoecke/go-spark"
	"strings"
)

type SparkEvent struct {
	Id          string `json:"id" binding:"required"`
	RoomId      string `json:"roomId" binding:"required"`
	PersonId    string `json:"personId" binding:"required"`
	PersonEmail string `json:"personEmail" binding:"required"`
	Text        string `json:"text" binding:"required"`
}

// quoteKey returns the key used for all quote entries.
func quoteKey(c context.Context) *datastore.Key {
	return datastore.NewKey(c, "SparkEvent", "default_spark_event", 0, nil)
}

func init() {

	m := martini.Classic()
	m.Use(render.Renderer())

	m.Use(func(res http.ResponseWriter, req *http.Request) {
		authorization := &spark.Authorization{AccessToken: os.Getenv("SPARK_TOKEN")}
		spark.InitClient(authorization)

		ctx := appengine.NewContext(req)
		spark.SetHttpClient(urlfetch.Client(ctx), ctx)
		log.Infof(ctx, "after setting http client, token = %s\n", os.Getenv("SPARK_TOKEN"))
	})

	m.Post("/spark", binding.Json(SparkEvent{}), func(sparkEvent SparkEvent, res http.ResponseWriter, req *http.Request, r render.Render) {
		ctx := appengine.NewContext(req)
		log.Infof(ctx, "Message = %s", sparkEvent.Text)

		if strings.HasPrefix(sparkEvent.Text, "/") {
			//person := spark.Person{ID: sparkEvent.PersonId}
			//person.Get()

			c := &Client{
				key:  os.Getenv("TRANSIT_511_KEY"),
				http: urlfetch.Client(ctx),
			}

			s := strings.Split(sparkEvent.Text, " ")
			command := s[0]
			log.Infof(ctx, "command = %s", command)
			if command == "/routes" {

				routeList := getRoutes(c)
				text := ""
				for _, route := range routeList.Route {
					text = text + route.Name + " (" + route.Code + ")\n"
				}

				message := spark.Message{
					RoomID: sparkEvent.RoomId,
					Text:   text,
				}
				message.Post()
			} else if command == "/stops" {

				routeList := getStops(c, s[1], s[2])

				log.Infof(ctx, "Route List = %v", routeList)

				text := ""
				for _, stop := range routeList.Route[0].RouteDirectionList[0].RouteDirection.StopList.Stop {
					text = text + stop.Name + " (" + stop.StopCode + ")\n"
				}

				message := spark.Message{
					RoomID: sparkEvent.RoomId,
					Text:   text,
				}
				message.Post()

			} else if command == "/departures" {

				routeList := getNextDepartures(c, s[1])

				log.Infof(ctx, "Route List = %v", routeList)

				text := ""
				for _, time := range routeList.Route[0].RouteDirectionList[0].RouteDirection.StopList.Stop[0].DepartureTimeList.DepartureTime {
					text = text + time + "\n"
				}

				message := spark.Message{
					RoomID: sparkEvent.RoomId,
					Text:   text,
				}
				message.Post()

			}

			key, err := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "SparkEvent", quoteKey(ctx)), &sparkEvent)
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}

			err1 := datastore.Get(ctx, key, &sparkEvent)
			if err1 != nil {
				http.Error(res, err1.Error(), http.StatusInternalServerError)
				return
			}

		}

	})

	m.Get("/events", func(res http.ResponseWriter, req *http.Request, r render.Render) {
		ctx := appengine.NewContext(req)

		q := datastore.NewQuery("SparkEvent").Ancestor(quoteKey(ctx)).Limit(10)

		events := make([]SparkEvent, 0, 10)
		q.GetAll(ctx, &events)
		r.JSON(200, events)
	})

	m.Get("/routes", func(res http.ResponseWriter, req *http.Request, r render.Render) {
		ctx := appengine.NewContext(req)
		c := &Client{
			key:  os.Getenv("TRANSIT_511_KEY"),
			http: urlfetch.Client(ctx),
		}

		routeList := getRoutes(c)
		r.JSON(200, routeList)
	})

	m.Get("/stops", func(res http.ResponseWriter, req *http.Request, r render.Render) {
		ctx := appengine.NewContext(req)
		c := &Client{
			key:  os.Getenv("TRANSIT_511_KEY"),
			http: urlfetch.Client(ctx),
		}

		routeList := getStops(c, "KT", "Inbound")
		r.JSON(200, routeList)
	})

	m.Get("/departures", func(res http.ResponseWriter, req *http.Request, r render.Render) {
		ctx := appengine.NewContext(req)
		c := &Client{
			key:  os.Getenv("TRANSIT_511_KEY"),
			http: urlfetch.Client(ctx),
		}

		routeList := getNextDepartures(c, "17356")
		r.JSON(200, routeList)
	})

	m.Get("/people", func(res http.ResponseWriter, req *http.Request, r render.Render) {
		people := spark.People{Displayname: "John O'Reilly"}
		people.Get()

		r.HTML(200, "people", people.Items)
	})

	http.Handle("/", m)
}
