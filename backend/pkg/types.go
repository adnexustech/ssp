package ssp

import (
	"time"
)

// Publisher represents a publisher entity
type Publisher struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Domain      string    `json:"domain"`
	Active      bool      `json:"active"`
	RevShare    float64   `json:"revShare"` // Publisher revenue share (0.0-1.0)
	PaymentInfo string    `json:"paymentInfo,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Site represents a publisher site
type Site struct {
	ID          string    `json:"id"`
	PublisherID string    `json:"publisherId"`
	Name        string    `json:"name"`
	Domain      string    `json:"domain"`
	Page        string    `json:"page,omitempty"`
	Cat         []string  `json:"cat,omitempty"` // IAB categories
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Placement represents an ad placement on a site
type Placement struct {
	ID          string         `json:"id"`
	SiteID      string         `json:"siteId"`
	Name        string         `json:"name"`
	AdType      string         `json:"adType"` // banner, video, native
	Width       int            `json:"width,omitempty"`
	Height      int            `json:"height,omitempty"`
	MinBidFloor float64        `json:"minBidFloor"`
	Active      bool           `json:"active"`
	Formats     []Format       `json:"formats,omitempty"` // For multi-size placements
	Video       *VideoSettings `json:"video,omitempty"`   // Video-specific settings
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

// Format represents an ad format size
type Format struct {
	W int `json:"w"`
	H int `json:"h"`
}

// VideoSettings represents video placement settings
type VideoSettings struct {
	Mimes          []string `json:"mimes"`
	MinDuration    int      `json:"minduration,omitempty"`
	MaxDuration    int      `json:"maxduration"`
	Protocols      []int    `json:"protocols"`
	Linearity      int      `json:"linearity,omitempty"`  // 1=linear, 2=non-linear
	StartDelay     int      `json:"startdelay,omitempty"` // -1=mid-roll, 0=pre-roll, >0=seconds
	PlaybackMethod []int    `json:"playbackmethod,omitempty"`
	API            []int    `json:"api,omitempty"` // Supported API frameworks
}

// OpenRTB 2.5 structures

// BidRequest represents an OpenRTB 2.5 bid request
type BidRequest struct {
	ID      string       `json:"id"`
	Imp     []Impression `json:"imp"`
	Site    *SiteInfo    `json:"site,omitempty"`
	App     *App         `json:"app,omitempty"`
	Device  *Device      `json:"device,omitempty"`
	User    *User        `json:"user,omitempty"`
	Test    int          `json:"test,omitempty"`
	At      int          `json:"at"`                // Auction type: 1=first price, 2=second price
	Tmax    int          `json:"tmax"`              // Max time in milliseconds
	WSeat   []string     `json:"wseat,omitempty"`   // Whitelist of buyer seats
	BSeat   []string     `json:"bseat,omitempty"`   // Blacklist of buyer seats
	AllImps int          `json:"allimps,omitempty"` // 1=all impressions in bid request
	Cur     []string     `json:"cur,omitempty"`
	WLang   []string     `json:"wlang,omitempty"` // Whitelist of languages
	BCat    []string     `json:"bcat,omitempty"`  // Blocked IAB categories
	BAdv    []string     `json:"badv,omitempty"`  // Blocked advertiser domains
	BApp    []string     `json:"bapp,omitempty"`  // Blocked app bundle/package names
	Source  *Source      `json:"source,omitempty"`
	Regs    *Regs        `json:"regs,omitempty"` // Regulations
	Ext     interface{}  `json:"ext,omitempty"`
}

// Impression represents a single ad impression
type Impression struct {
	ID                string      `json:"id"`
	Banner            *Banner     `json:"banner,omitempty"`
	Video             *Video      `json:"video,omitempty"`
	Audio             *Audio      `json:"audio,omitempty"`
	Native            *Native     `json:"native,omitempty"`
	PMP               *PMP        `json:"pmp,omitempty"` // Private marketplace
	DisplayManager    string      `json:"displaymanager,omitempty"`
	DisplayManagerVer string      `json:"displaymanagerver,omitempty"`
	Instl             int         `json:"instl,omitempty"` // 1=interstitial
	TagID             string      `json:"tagid,omitempty"` // Placement tag ID
	BidFloor          float64     `json:"bidfloor,omitempty"`
	BidFloorCur       string      `json:"bidfloorcur,omitempty"`
	Clickbrowser      int         `json:"clickbrowser,omitempty"`
	Secure            int         `json:"secure,omitempty"` // 1=requires HTTPS
	IFrameBuster      []string    `json:"iframebuster,omitempty"`
	Exp               int         `json:"exp,omitempty"` // Advisory as to number of seconds that may elapse
	Ext               interface{} `json:"ext,omitempty"`
}

// Banner represents banner ad specifications
type Banner struct {
	Format   []Format    `json:"format,omitempty"` // Array of format objects
	W        int         `json:"w,omitempty"`
	H        int         `json:"h,omitempty"`
	WMax     int         `json:"wmax,omitempty"`
	WMin     int         `json:"wmin,omitempty"`
	HMax     int         `json:"hmax,omitempty"`
	HMin     int         `json:"hmin,omitempty"`
	BType    []int       `json:"btype,omitempty"` // Blocked creative types
	BAttr    []int       `json:"battr,omitempty"` // Blocked creative attributes
	Pos      int         `json:"pos,omitempty"`   // Ad position
	MIMEs    []string    `json:"mimes,omitempty"`
	TopFrame int         `json:"topframe,omitempty"` // 1=in top frame
	ExpDir   []int       `json:"expdir,omitempty"`   // Expandable ad directions
	API      []int       `json:"api,omitempty"`      // Supported API frameworks
	ID       string      `json:"id,omitempty"`
	VCM      int         `json:"vcm,omitempty"` // Video companion ad mode
	Ext      interface{} `json:"ext,omitempty"`
}

// Video represents video ad specifications (OpenRTB 2.5)
type Video struct {
	Mimes          []string    `json:"mimes"`
	MinDuration    int         `json:"minduration,omitempty"`
	MaxDuration    int         `json:"maxduration"`
	Protocols      []int       `json:"protocols"`
	Protocol       int         `json:"protocol,omitempty"` // Deprecated in favor of protocols
	W              int         `json:"w,omitempty"`
	H              int         `json:"h,omitempty"`
	StartDelay     int         `json:"startdelay,omitempty"`
	Placement      int         `json:"placement,omitempty"` // Video placement type
	Linearity      int         `json:"linearity,omitempty"`
	Skip           int         `json:"skip,omitempty"`      // 1=skippable
	SkipMin        int         `json:"skipmin,omitempty"`   // Seconds before skip button
	SkipAfter      int         `json:"skipafter,omitempty"` // Seconds video must play
	Sequence       int         `json:"sequence,omitempty"`
	BAttr          []int       `json:"battr,omitempty"`
	MaxExtended    int         `json:"maxextended,omitempty"`
	MinBitrate     int         `json:"minbitrate,omitempty"`
	MaxBitrate     int         `json:"maxbitrate,omitempty"`
	BoxingAllowed  int         `json:"boxingallowed,omitempty"` // 1=allowed
	PlaybackMethod []int       `json:"playbackmethod,omitempty"`
	PlaybackEnd    int         `json:"playbackend,omitempty"` // Playback termination mode
	Delivery       []int       `json:"delivery,omitempty"`
	Pos            int         `json:"pos,omitempty"`
	CompanionAd    []Banner    `json:"companionad,omitempty"`
	API            []int       `json:"api,omitempty"`
	CompanionType  []int       `json:"companiontype,omitempty"`
	Ext            interface{} `json:"ext,omitempty"`
}

// Audio represents audio ad specifications
type Audio struct {
	Mimes         []string    `json:"mimes"`
	MinDuration   int         `json:"minduration,omitempty"`
	MaxDuration   int         `json:"maxduration"`
	Protocols     []int       `json:"protocols"`
	StartDelay    int         `json:"startdelay,omitempty"`
	Sequence      int         `json:"sequence,omitempty"`
	BAttr         []int       `json:"battr,omitempty"`
	MaxExtended   int         `json:"maxextended,omitempty"`
	MinBitrate    int         `json:"minbitrate,omitempty"`
	MaxBitrate    int         `json:"maxbitrate,omitempty"`
	Delivery      []int       `json:"delivery,omitempty"`
	CompanionAd   []Banner    `json:"companionad,omitempty"`
	API           []int       `json:"api,omitempty"`
	CompanionType []int       `json:"companiontype,omitempty"`
	MaxSeq        int         `json:"maxseq,omitempty"`
	Feed          int         `json:"feed,omitempty"`
	Stitched      int         `json:"stitched,omitempty"`
	NVol          int         `json:"nvol,omitempty"` // Volume normalization mode
	Ext           interface{} `json:"ext,omitempty"`
}

// Native represents native ad specifications
type Native struct {
	Request string      `json:"request"` // JSON-encoded Native request
	Ver     string      `json:"ver,omitempty"`
	API     []int       `json:"api,omitempty"`
	BAttr   []int       `json:"battr,omitempty"`
	Ext     interface{} `json:"ext,omitempty"`
}

// SiteInfo represents publisher site information in bid request
type SiteInfo struct {
	ID            string      `json:"id,omitempty"`
	Name          string      `json:"name,omitempty"`
	Domain        string      `json:"domain,omitempty"`
	Cat           []string    `json:"cat,omitempty"`
	SectionCat    []string    `json:"sectioncat,omitempty"`
	PageCat       []string    `json:"pagecat,omitempty"`
	Page          string      `json:"page,omitempty"`
	Ref           string      `json:"ref,omitempty"`
	Search        string      `json:"search,omitempty"`
	Mobile        int         `json:"mobile,omitempty"`
	PrivacyPolicy int         `json:"privacypolicy,omitempty"`
	Publisher     *Publisher2 `json:"publisher,omitempty"`
	Content       *Content    `json:"content,omitempty"`
	Keywords      string      `json:"keywords,omitempty"`
	Ext           interface{} `json:"ext,omitempty"`
}

// App represents mobile app information
type App struct {
	ID            string      `json:"id,omitempty"`
	Name          string      `json:"name,omitempty"`
	Bundle        string      `json:"bundle,omitempty"`
	Domain        string      `json:"domain,omitempty"`
	StoreURL      string      `json:"storeurl,omitempty"`
	Cat           []string    `json:"cat,omitempty"`
	SectionCat    []string    `json:"sectioncat,omitempty"`
	PageCat       []string    `json:"pagecat,omitempty"`
	Ver           string      `json:"ver,omitempty"`
	PrivacyPolicy int         `json:"privacypolicy,omitempty"`
	Paid          int         `json:"paid,omitempty"`
	Publisher     *Publisher2 `json:"publisher,omitempty"`
	Content       *Content    `json:"content,omitempty"`
	Keywords      string      `json:"keywords,omitempty"`
	Ext           interface{} `json:"ext,omitempty"`
}

// Publisher2 represents publisher information in bid request (named to avoid conflict)
type Publisher2 struct {
	ID     string      `json:"id,omitempty"`
	Name   string      `json:"name,omitempty"`
	Cat    []string    `json:"cat,omitempty"`
	Domain string      `json:"domain,omitempty"`
	Ext    interface{} `json:"ext,omitempty"`
}

// Content represents content metadata
type Content struct {
	ID                 string      `json:"id,omitempty"`
	Episode            int         `json:"episode,omitempty"`
	Title              string      `json:"title,omitempty"`
	Series             string      `json:"series,omitempty"`
	Season             string      `json:"season,omitempty"`
	Artist             string      `json:"artist,omitempty"`
	Genre              string      `json:"genre,omitempty"`
	Album              string      `json:"album,omitempty"`
	ISRC               string      `json:"isrc,omitempty"`
	Producer           *Producer   `json:"producer,omitempty"`
	URL                string      `json:"url,omitempty"`
	Cat                []string    `json:"cat,omitempty"`
	ProdQ              int         `json:"prodq,omitempty"`   // Production quality
	Context            int         `json:"context,omitempty"` // Content context
	ContentRating      string      `json:"contentrating,omitempty"`
	UserRating         string      `json:"userrating,omitempty"`
	QAGMediaRating     int         `json:"qagmediarating,omitempty"`
	Keywords           string      `json:"keywords,omitempty"`
	LiveStream         int         `json:"livestream,omitempty"`
	SourceRelationship int         `json:"sourcerelationship,omitempty"`
	Len                int         `json:"len,omitempty"` // Content length in seconds
	Language           string      `json:"language,omitempty"`
	Embeddable         int         `json:"embeddable,omitempty"`
	Data               []Data      `json:"data,omitempty"`
	Ext                interface{} `json:"ext,omitempty"`
}

// Producer represents content producer
type Producer struct {
	ID     string      `json:"id,omitempty"`
	Name   string      `json:"name,omitempty"`
	Cat    []string    `json:"cat,omitempty"`
	Domain string      `json:"domain,omitempty"`
	Ext    interface{} `json:"ext,omitempty"`
}

// Device represents user device information
type Device struct {
	UA             string      `json:"ua,omitempty"` // User agent
	Geo            *Geo        `json:"geo,omitempty"`
	DNT            int         `json:"dnt,omitempty"` // Do Not Track: 0=tracking allowed, 1=tracking not allowed
	Lmt            int         `json:"lmt,omitempty"` // Limit Ad Tracking
	IP             string      `json:"ip,omitempty"`
	IPv6           string      `json:"ipv6,omitempty"`
	DeviceType     int         `json:"devicetype,omitempty"` // Device type from OpenRTB spec
	Make           string      `json:"make,omitempty"`
	Model          string      `json:"model,omitempty"`
	OS             string      `json:"os,omitempty"`
	OSV            string      `json:"osv,omitempty"`
	HWV            string      `json:"hwv,omitempty"`      // Hardware version
	H              int         `json:"h,omitempty"`        // Screen height
	W              int         `json:"w,omitempty"`        // Screen width
	PPI            int         `json:"ppi,omitempty"`      // Pixels per inch
	PxRatio        float64     `json:"pxratio,omitempty"`  // Physical pixel ratio
	JS             int         `json:"js,omitempty"`       // JavaScript support
	GeoFetch       int         `json:"geofetch,omitempty"` // Indicates if geolocation API will be available
	FlashVer       string      `json:"flashver,omitempty"`
	Language       string      `json:"language,omitempty"`
	Carrier        string      `json:"carrier,omitempty"`
	MCCMNC         string      `json:"mccmnc,omitempty"`
	ConnectionType int         `json:"connectiontype,omitempty"`
	IFA            string      `json:"ifa,omitempty"`      // ID for Advertisers
	DIDSHA1        string      `json:"didsha1,omitempty"`  // Device ID SHA1
	DIDMD5         string      `json:"didmd5,omitempty"`   // Device ID MD5
	DPIDSHA1       string      `json:"dpidsha1,omitempty"` // Platform Device ID SHA1
	DPIDMD5        string      `json:"dpidmd5,omitempty"`  // Platform Device ID MD5
	MacSHA1        string      `json:"macsha1,omitempty"`
	MacMD5         string      `json:"macmd5,omitempty"`
	Ext            interface{} `json:"ext,omitempty"`
}

// Geo represents geographic location
type Geo struct {
	Lat           float64     `json:"lat,omitempty"`
	Lon           float64     `json:"lon,omitempty"`
	Type          int         `json:"type,omitempty"` // Location source: 1=GPS, 2=IP, 3=user provided
	Accuracy      int         `json:"accuracy,omitempty"`
	LastFix       int         `json:"lastfix,omitempty"`   // Seconds since last geolocation fix
	IPService     int         `json:"ipservice,omitempty"` // IP location service
	Country       string      `json:"country,omitempty"`   // ISO-3166-1 Alpha-3
	Region        string      `json:"region,omitempty"`    // ISO-3166-2
	RegionFIPS104 string      `json:"regionfips104,omitempty"`
	Metro         string      `json:"metro,omitempty"`
	City          string      `json:"city,omitempty"`
	ZIP           string      `json:"zip,omitempty"`
	UTC           int         `json:"utcoffset,omitempty"` // UTC offset in minutes
	Ext           interface{} `json:"ext,omitempty"`
}

// User represents user information
type User struct {
	ID         string      `json:"id,omitempty"`
	BuyerUID   string      `json:"buyeruid,omitempty"`
	YOB        int         `json:"yob,omitempty"` // Year of birth
	Gender     string      `json:"gender,omitempty"`
	Keywords   string      `json:"keywords,omitempty"`
	CustomData string      `json:"customdata,omitempty"`
	Geo        *Geo        `json:"geo,omitempty"`
	Data       []Data      `json:"data,omitempty"`
	Consent    string      `json:"consent,omitempty"` // GDPR consent string
	Ext        interface{} `json:"ext,omitempty"`
}

// Data represents additional data segments
type Data struct {
	ID      string      `json:"id,omitempty"`
	Name    string      `json:"name,omitempty"`
	Segment []Segment   `json:"segment,omitempty"`
	Ext     interface{} `json:"ext,omitempty"`
}

// Segment represents a data segment
type Segment struct {
	ID    string      `json:"id,omitempty"`
	Name  string      `json:"name,omitempty"`
	Value string      `json:"value,omitempty"`
	Ext   interface{} `json:"ext,omitempty"`
}

// PMP represents private marketplace
type PMP struct {
	PrivateAuction int         `json:"private_auction,omitempty"` // 1=private auction
	Deals          []Deal      `json:"deals,omitempty"`
	Ext            interface{} `json:"ext,omitempty"`
}

// Deal represents a private deal
type Deal struct {
	ID          string      `json:"id"`
	BidFloor    float64     `json:"bidfloor,omitempty"`
	BidFloorCur string      `json:"bidfloorcur,omitempty"`
	At          int         `json:"at,omitempty"`       // Auction type
	WSeat       []string    `json:"wseat,omitempty"`    // Whitelist of buyer seats
	WADomain    []string    `json:"wadomain,omitempty"` // Whitelist of advertiser domains
	Ext         interface{} `json:"ext,omitempty"`
}

// Source represents the source of the bid request
type Source struct {
	FD     int         `json:"fd,omitempty"`     // 1=final destination, 0=intermediary
	TID    string      `json:"tid,omitempty"`    // Transaction ID
	PChain string      `json:"pchain,omitempty"` // Payment chain
	Ext    interface{} `json:"ext,omitempty"`
}

// Regs represents regulatory information
type Regs struct {
	Coppa int         `json:"coppa,omitempty"` // 1=subject to COPPA
	Ext   interface{} `json:"ext,omitempty"`   // Contains GDPR flag
}

// BidResponse represents an OpenRTB 2.5 bid response
type BidResponse struct {
	ID         string      `json:"id"`
	SeatBid    []SeatBid   `json:"seatbid,omitempty"`
	BidID      string      `json:"bidid,omitempty"`
	Cur        string      `json:"cur,omitempty"`
	CustomData string      `json:"customdata,omitempty"`
	NBR        int         `json:"nbr,omitempty"` // No-bid reason code
	Ext        interface{} `json:"ext,omitempty"`
}

// SeatBid represents bids from a seat
type SeatBid struct {
	Bid   []Bid       `json:"bid"`
	Seat  string      `json:"seat,omitempty"`
	Group int         `json:"group,omitempty"` // 1=impressions must be won together
	Ext   interface{} `json:"ext,omitempty"`
}

// Bid represents a single bid
type Bid struct {
	ID             string      `json:"id"`
	ImpID          string      `json:"impid"`
	Price          float64     `json:"price"`
	AdID           string      `json:"adid,omitempty"`
	NURL           string      `json:"nurl,omitempty"` // Win notice URL
	BURL           string      `json:"burl,omitempty"` // Billing notice URL
	LURL           string      `json:"lurl,omitempty"` // Loss notice URL
	ADM            string      `json:"adm,omitempty"`  // Ad markup
	ADomain        []string    `json:"adomain,omitempty"`
	Bundle         string      `json:"bundle,omitempty"`
	IURL           string      `json:"iurl,omitempty"` // Image URL for content checking
	CID            string      `json:"cid,omitempty"`  // Campaign ID
	CRID           string      `json:"crid,omitempty"` // Creative ID
	Tactic         string      `json:"tactic,omitempty"`
	Cat            []string    `json:"cat,omitempty"`  // IAB categories
	Attr           []int       `json:"attr,omitempty"` // Creative attributes
	API            int         `json:"api,omitempty"`
	Protocol       int         `json:"protocol,omitempty"`
	QAGMediaRating int         `json:"qagmediarating,omitempty"`
	Language       string      `json:"language,omitempty"`
	DealID         string      `json:"dealid,omitempty"`
	W              int         `json:"w,omitempty"`
	H              int         `json:"h,omitempty"`
	WRatio         int         `json:"wratio,omitempty"` // Relative width for native
	HRatio         int         `json:"hratio,omitempty"` // Relative height for native
	Exp            int         `json:"exp,omitempty"`    // Expiry time
	Ext            interface{} `json:"ext,omitempty"`
}

// SupplyStats represents supply-side statistics
type SupplyStats struct {
	PublisherID string  `json:"publisherId"`
	SiteID      string  `json:"siteId"`
	PlacementID string  `json:"placementId"`
	Requests    int64   `json:"requests"`
	Impressions int64   `json:"impressions"`
	Revenue     float64 `json:"revenue"`
	Fills       int64   `json:"fills"`
	AvgCPM      float64 `json:"avgCpm"`
	Date        string  `json:"date"`
}

// AdTag represents generated ad tag code
type AdTag struct {
	PlacementID string    `json:"placementId"`
	TagType     string    `json:"tagType"` // display, video, header-bidding
	HTML        string    `json:"html"`
	JavaScript  string    `json:"javascript"`
	CreatedAt   time.Time `json:"createdAt"`
}
