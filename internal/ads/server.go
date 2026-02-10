package ads

import (
	realip "github.com/Ferluci/fast-realip"
	"github.com/VyacheslavGoryunov/simple-ads-server/internal/stats"
	"github.com/mssola/user_agent"
	"github.com/oschwald/geoip2-golang"
	"github.com/valyala/fasthttp"
	"log"
	"net"
	"net/http"
	"time"
)

type Server struct {
	geoip *geoip2.Reader
	stats *stats.Manager
}

func NewServer(geoip *geoip2.Reader, stats *stats.Manager) *Server {
	return &Server{geoip: geoip, stats: stats}
}

func (s *Server) Listen() error {
	return fasthttp.ListenAndServe(":8080", s.handler)
}

func (s *Server) handler(ctx *fasthttp.RequestCtx) {
	start := time.Now()
	defer func() {
		observeRequest(time.Since(start), ctx.Response.StatusCode())
	}()
	remoteIp := realip.FromRequest(ctx)
	ua := string(ctx.Request.Header.UserAgent())

	parsed := user_agent.New(ua)
	browserName, _ := parsed.Browser()
	osName := parsed.OS()

	countryCode := "ZZ"
	country, err := s.geoip.Country(net.ParseIP(remoteIp))
	if err != nil {
		log.Printf("Failed to parse country: %v", err)
	} else {
		countryCode = country.Country.IsoCode
	}

	user := &User{
		Country: countryCode,
		Browser: browserName,
	}
	campaigns := GetStaticCampaigns()

	winner := MakeAuction(campaigns, user)
	key := stats.NewKey(stats.Key{
		Country:    user.Country,
		Os:         osName,
		Browser:    browserName,
		CampaignId: 0,
	})
	value := stats.Value{
		Requests:    1,
		Impressions: 0,
	}
	if winner != nil {
		key.CampaignId = winner.ID
		value.Impressions = 1
	}
	s.stats.Append(key, value)
	if winner == nil {
		ctx.Redirect("https://example.com", http.StatusSeeOther)
		return
	}

	ctx.Redirect(winner.ClickUrl, http.StatusSeeOther)
}
