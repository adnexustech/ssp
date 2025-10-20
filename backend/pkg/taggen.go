package ssp

import (
	"fmt"
	"html/template"
)

// TagGenerator generates ad tags for publishers
type TagGenerator struct {
	sspEndpoint string
	cdnURL      string
}

// NewTagGenerator creates a new tag generator
func NewTagGenerator(sspEndpoint, cdnURL string) *TagGenerator {
	return &TagGenerator{
		sspEndpoint: sspEndpoint,
		cdnURL:      cdnURL,
	}
}

// GenerateDisplayTag generates a display ad tag
func (tg *TagGenerator) GenerateDisplayTag(placement *Placement) (string, error) {
	tmpl := `<!-- AdNexus SSP Display Ad Tag -->
<div id="adnexus-{{.PlacementID}}" style="width:{{.Width}}px;height:{{.Height}}px;"></div>
<script>
(function() {
  var adnexus = window.adnexus || {};
  adnexus.placements = adnexus.placements || [];
  adnexus.placements.push({
    placementId: '{{.PlacementID}}',
    width: {{.Width}},
    height: {{.Height}},
    endpoint: '{{.SSPEndpoint}}/ad/request'
  });

  if (!window.adnexusLoaded) {
    var s = document.createElement('script');
    s.async = true;
    s.src = '{{.CDNURL}}/adnexus-ssp.js';
    document.head.appendChild(s);
    window.adnexusLoaded = true;
  }
})();
</script>`

	t, err := template.New("display").Parse(tmpl)
	if err != nil {
		return "", err
	}

	data := struct {
		PlacementID string
		Width       int
		Height      int
		SSPEndpoint string
		CDNURL      string
	}{
		PlacementID: placement.ID,
		Width:       placement.Width,
		Height:      placement.Height,
		SSPEndpoint: tg.sspEndpoint,
		CDNURL:      tg.cdnURL,
	}

	var buf []byte
	w := &writeBuffer{buf: buf}
	if err := t.Execute(w, data); err != nil {
		return "", err
	}

	return string(w.buf), nil
}

// GenerateVASTTag generates a VAST video ad tag
func (tg *TagGenerator) GenerateVASTTag(placement *Placement) (string, error) {
	tmpl := `<!-- AdNexus SSP VAST Video Ad Tag -->
<div id="adnexus-video-{{.PlacementID}}"></div>
<script>
(function() {
  var adnexus = window.adnexus || {};
  adnexus.videoPlacements = adnexus.videoPlacements || [];
  adnexus.videoPlacements.push({
    placementId: '{{.PlacementID}}',
    width: {{.Width}},
    height: {{.Height}},
    vastUrl: '{{.SSPEndpoint}}/vast/{{.PlacementID}}',
    containerId: 'adnexus-video-{{.PlacementID}}'
  });

  if (!window.adnexusVideoLoaded) {
    var s = document.createElement('script');
    s.async = true;
    s.src = '{{.CDNURL}}/adnexus-video.js';
    document.head.appendChild(s);
    window.adnexusVideoLoaded = true;
  }
})();
</script>`

	t, err := template.New("vast").Parse(tmpl)
	if err != nil {
		return "", err
	}

	data := struct {
		PlacementID string
		Width       int
		Height      int
		SSPEndpoint string
		CDNURL      string
	}{
		PlacementID: placement.ID,
		Width:       placement.Width,
		Height:      placement.Height,
		SSPEndpoint: tg.sspEndpoint,
		CDNURL:      tg.cdnURL,
	}

	var buf []byte
	w := &writeBuffer{buf: buf}
	if err := t.Execute(w, data); err != nil {
		return "", err
	}

	return string(w.buf), nil
}

// GenerateHeaderBiddingTag generates a Prebid.js compatible header bidding tag
func (tg *TagGenerator) GenerateHeaderBiddingTag(placement *Placement) (string, error) {
	tmpl := `<!-- AdNexus SSP Header Bidding Tag -->
<script>
var adnexusPrebid = adnexusPrebid || {};
adnexusPrebid.que = adnexusPrebid.que || [];

adnexusPrebid.que.push(function() {
  var adUnits = [{
    code: 'adnexus-hb-{{.PlacementID}}',
    mediaTypes: {
      banner: {
        sizes: [{{.Sizes}}]
      }
    },
    bids: [{
      bidder: 'adnexus',
      params: {
        placementId: '{{.PlacementID}}',
        endpoint: '{{.SSPEndpoint}}/openrtb2/auction'
      }
    }]
  }];

  adnexusPrebid.addAdUnits(adUnits);
  adnexusPrebid.requestBids({
    bidsBackHandler: function() {
      adnexusPrebid.setTargetingForGPTAsync(['adnexus-hb-{{.PlacementID}}']);
    }
  });
});
</script>

<div id='adnexus-hb-{{.PlacementID}}' style='width:{{.Width}}px;height:{{.Height}}px;'>
  <script>
    googletag.cmd.push(function() {
      googletag.display('adnexus-hb-{{.PlacementID}}');
    });
  </script>
</div>`

	t, err := template.New("headerbidding").Parse(tmpl)
	if err != nil {
		return "", err
	}

	// Build sizes string
	sizes := ""
	if len(placement.Formats) > 0 {
		for i, f := range placement.Formats {
			if i > 0 {
				sizes += ", "
			}
			sizes += fmt.Sprintf("[%d, %d]", f.W, f.H)
		}
	} else {
		sizes = fmt.Sprintf("[%d, %d]", placement.Width, placement.Height)
	}

	data := struct {
		PlacementID string
		Width       int
		Height      int
		Sizes       string
		SSPEndpoint string
	}{
		PlacementID: placement.ID,
		Width:       placement.Width,
		Height:      placement.Height,
		Sizes:       sizes,
		SSPEndpoint: tg.sspEndpoint,
	}

	var buf []byte
	w := &writeBuffer{buf: buf}
	if err := t.Execute(w, data); err != nil {
		return "", err
	}

	return string(w.buf), nil
}

// GenerateVASTXML generates a VAST XML response
func (tg *TagGenerator) GenerateVASTXML(bid *Bid, placement *Placement) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<VAST version="3.0">
  <Ad id="%s">
    <InLine>
      <AdSystem>AdNexus SSP</AdSystem>
      <AdTitle>Advertisement</AdTitle>
      <Impression><![CDATA[%s]]></Impression>
      <Creatives>
        <Creative>
          <Linear>
            <Duration>00:00:30</Duration>
            <MediaFiles>
              <MediaFile delivery="progressive" type="video/mp4" width="%d" height="%d">
                <![CDATA[%s]]>
              </MediaFile>
            </MediaFiles>
            <VideoClicks>
              <ClickThrough><![CDATA[%s]]></ClickThrough>
            </VideoClicks>
          </Linear>
        </Creative>
      </Creatives>
    </InLine>
  </Ad>
</VAST>`, bid.ID, bid.NURL, placement.Width, placement.Height, bid.IURL, bid.ADM)
}

// writeBuffer implements io.Writer for template execution
type writeBuffer struct {
	buf []byte
}

func (w *writeBuffer) Write(p []byte) (n int, err error) {
	w.buf = append(w.buf, p...)
	return len(p), nil
}
