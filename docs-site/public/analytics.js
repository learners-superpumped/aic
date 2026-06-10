/*
 * aic CLI docs (docs.runaic.com) — Amplitude + PostHog 듀얼 계측
 *
 * aicompany 공용 Amplitude(AINATIVE) + 공용 PostHog 프로젝트로 동시 전송하고,
 * 모든 이벤트에 service="runaic-docs" 를 붙여 다른 서비스와 구분한다.
 * (runaic.com 랜딩은 service="runaic-landing" — koreaaitel/clawops 와 동일 컨벤션.)
 *
 * 정적 빌드 산출물에 그대로 포함되는 파일이라 공개 클라이언트 키를 직접 둔다 —
 * 브라우저 ingest 키는 NEXT_PUBLIC_* 네이밍대로 설계상 공개값이다.
 * (비밀 키는 별도의 AMPLITUDE_SECRET_KEY / POSTHOG personal key.)
 *
 * autocapture 가 페이지뷰·요소 클릭·세션·유입경로(attribution)를 자동 수집하고,
 * 아래 위임 리스너가 설치 명령 복사·GitHub 이탈을 의미 있는 퍼널 이벤트로 보강한다.
 */
(function () {
  "use strict";

  var AMPLITUDE_API_KEY = "41d49d9b6fd3ed820611220be9c39288";
  var POSTHOG_API_KEY = "phc_uZEkLXSxGC45F4JV6ChosojY5FNUnkKogNW7tiMTcfsi";
  var POSTHOG_HOST = "https://us.i.posthog.com";
  var SERVICE = "runaic-docs";

  /* 링크 프리뷰·AI 크롤러 봇은 init 자체를 skip — page_view/attribution 오염 방지.
     (clawops-landing 과 동일 컨벤션) */
  var BOT_UA_PATTERN = /bot|crawl(er)?|spider|slurp|mediapartners|chatgpt|claudebot|gptbot|ccbot|anthropic|perplexity|ai2bot|bytespider|applebot|yandex|naverbot|daum|facebookexternalhit|twitterbot|linkedinbot|whatsapp|telegram|discord|pinterest|teams|slackbot|prerender|headlesschrome|phantomjs|puppeteer|playwright/i;
  if (navigator.webdriver || BOT_UA_PATTERN.test(navigator.userAgent || "")) return;

  /* --- Amplitude Browser SDK 2 공식 로더 스니펫 (v2.42.3) --- */
  !function(){"use strict";!function(e,t){var r=e.amplitude||{_q:[],_iq:{}};if(r.invoked)e.console&&console.error&&console.error("Amplitude snippet has been loaded.");else{var n=function(e,t){e.prototype[t]=function(){return this._q.push({name:t,args:Array.prototype.slice.call(arguments,0)}),this}},s=function(e,t,r){return function(n){e._q.push({name:t,args:Array.prototype.slice.call(r,0),resolve:n})}},o=function(e,t,r){e[t]=function(){if(r)return{promise:new Promise(s(e,t,Array.prototype.slice.call(arguments)))};!function(e,t,r){e._q.push({name:t,args:Array.prototype.slice.call(r,0)})}(e,t,Array.prototype.slice.call(arguments))}},i=function(e){for(var t=0;t<y.length;t++)o(e,y[t],!1);for(var r=0;r<g.length;r++)o(e,g[r],!0)};r.invoked=!0;var c=t.createElement("script");c.type="text/javascript",c.integrity="sha384-hmnfWlRQ5S6H/wCNnDtjTifF4/Jpkj4yezjl1CxwuhX54oCx83KDHuOdZKVIkdrR",c.crossOrigin="anonymous",c.async=!0,c.src="https://cdn.amplitude.com/libs/analytics-browser-2.42.3-min.js.gz",c.onload=function(){e.amplitude.runQueuedFunctions||console.log("[Amplitude] Error: could not load SDK")};var a=t.getElementsByTagName("script")[0];a.parentNode.insertBefore(c,a);for(var u=function(){return this._q=[],this},l=["add","append","clearAll","prepend","set","setOnce","unset","preInsert","postInsert","remove","getUserProperties"],p=0;p<l.length;p++)n(u,l[p]);r.Identify=u;for(var d=function(){return this._q=[],this},f=["getEventProperties","setProductId","setQuantity","setPrice","setRevenue","setRevenueType","setReceipt","setReceiptSig","setCurrency","setEventProperties"],v=0;v<f.length;v++)n(d,f[v]);r.Revenue=d;var y=["getDeviceId","setDeviceId","getSessionId","setSessionId","getUserId","setUserId","setOptOut","setTransport","reset","extendSession"],g=["init","add","remove","track","logEvent","identify","groupIdentify","setGroup","revenue","flush"];i(r),r.createInstance=function(e){return r._iq[e]={_q:[]},i(r._iq[e]),r._iq[e]},e.amplitude=r}}(window,document)}();

  var amplitude = window.amplitude;

  /* 모든 이벤트에 service 태그 자동 주입. init 전에 등록해 첫 page view 도 포함. */
  amplitude.add({
    name: "service-tag",
    type: "enrichment",
    setup: function () {
      return Promise.resolve();
    },
    execute: function (event) {
      event.event_properties = Object.assign({}, event.event_properties, {
        service: SERVICE,
      });
      return Promise.resolve(event);
    },
  });

  amplitude.init(AMPLITUDE_API_KEY, undefined, {
    autocapture: {
      attribution: true, // 유입경로: utm_*·referrer·gclid·fbclid 자동 수집
      pageViews: true, // 어떤 문서 페이지를 방문했는지
      sessions: true,
      elementInteractions: true, // 어디를 클릭했는지 (모든 요소)
      formInteractions: true,
      fileDownloads: true,
    },
    // runaic.com 랜딩과 device-id 를 공유해 landing→docs 유입을 한 퍼널로.
    cookieOptions: { domain: ".runaic.com" },
  });

  /* --- PostHog 공식 array.js 로더 스니펫 (autocapture + pageview 기본 ON) --- */
  !function(t,e){var o,n,p,r;e.__SV||(window.posthog && window.posthog.__loaded)||(window.posthog=e,e._i=[],e.init=function(i,s,a){function g(t,e){var o=e.split(".");2==o.length&&(t=t[o[0]],e=o[1]),t[e]=function(){t.push([e].concat(Array.prototype.slice.call(arguments,0)))}}(p=t.createElement("script")).type="text/javascript",p.crossOrigin="anonymous",p.async=!0,p.src=s.api_host.replace(".i.posthog.com","-assets.i.posthog.com")+"/static/array.js",(r=t.getElementsByTagName("script")[0]).parentNode.insertBefore(p,r);var u=e;for(void 0!==a?u=e[a]=[]:a="posthog",u.people=u.people||[],u.toString=function(t){var e="posthog";return"posthog"!==a&&(e+="."+a),t||(e+=" (stub)"),e},u.people.toString=function(){return u.toString(1)+".people (stub)"},o="Mi Ri init Vi Gi Rr Wi Ji Bi capture calculateEventProperties tn register register_once register_for_session unregister unregister_for_session an getFeatureFlag getFeatureFlagPayload getFeatureFlagResult isFeatureEnabled reloadFeatureFlags updateFlags updateEarlyAccessFeatureEnrollment getEarlyAccessFeatures on onFeatureFlags onSurveysLoaded onSessionId getSurveys getActiveMatchingSurveys renderSurvey displaySurvey cancelPendingSurvey canRenderSurvey canRenderSurveyAsync un identify setPersonProperties group resetGroups setPersonPropertiesForFlags resetPersonPropertiesForFlags setGroupPropertiesForFlags resetGroupPropertiesForFlags reset setIdentity clearIdentity get_distinct_id getGroups get_session_id get_session_replay_url alias set_config startSessionRecording stopSessionRecording sessionRecordingStarted captureException addExceptionStep captureLog startExceptionAutocapture stopExceptionAutocapture loadToolbar get_property getSessionProperty nn Xi createPersonProfile setInternalOrTestUser sn Hi cn opt_in_capturing opt_out_capturing has_opted_in_capturing has_opted_out_capturing get_explicit_consent_status is_capturing clear_opt_in_out_capturing Ki debug Lr rn getPageViewId captureTraceFeedback captureTraceMetric Di".split(" "),n=0;n<o.length;n++)g(u,o[n]);e._i.push([i,s,a])},e.__SV=1)}(document,window.posthog||[]);

  window.posthog.init(POSTHOG_API_KEY, {
    api_host: POSTHOG_HOST,
    defaults: "2026-01-30",
    person_profiles: "identified_only",
    autocapture: true,
    capture_pageview: true,
    cross_subdomain_cookie: true,
  });
  window.posthog.register({ service: SERVICE });
  window.posthog.group("service", SERVICE);

  function track(name, props) {
    var payload = Object.assign({ service: SERVICE }, props || {});
    try {
      amplitude.track(name, payload);
    } catch (e) {
      /* 분석 실패가 사이트를 깨뜨리지 않게 한다 */
    }
    try {
      window.posthog.capture(name, payload);
    } catch (e) {
      /* swallow */
    }
  }

  /* --- landed_from_social: UTM 유입 시 1건. ops-social 카드뉴스/캡션과 join --- */
  (function () {
    try {
      var p = new URLSearchParams(window.location.search);
      var src = p.get("utm_source"),
        med = p.get("utm_medium"),
        camp = p.get("utm_campaign");
      if (!src && !med && !camp) return;
      var refDomain = null;
      try {
        refDomain = document.referrer ? new URL(document.referrer).host : null;
      } catch (e) {}
      track("landed_from_social", {
        utm_source: src,
        utm_medium: med,
        utm_campaign: camp, // ops-social idempotencyKey 와 동일 슬러그
        utm_content: p.get("utm_content"),
        utm_term: p.get("utm_term"),
        referrer: document.referrer || null,
        referring_domain: refDomain,
        landing_path: window.location.pathname,
      });
    } catch (e) {}
  })();

  /* --- 설치 명령 복사·GitHub 이탈을 의미 있는 퍼널 이벤트로 보강 ---
     Starlight 는 MPA 라 페이지마다 이 파일이 다시 실행된다. */
  document.addEventListener(
    "click",
    function (e) {
      var node = e.target;
      if (!node || !node.closest) return;

      /* Expressive Code 코드블록의 copy 버튼 — 설치/명령 복사 퍼널 */
      var copyBtn = node.closest(".expressive-code .copy button, [data-code]");
      if (copyBtn) {
        track("docs_code_copied", {
          code: (copyBtn.getAttribute("data-code") || "").slice(0, 120) || null,
          page: window.location.pathname,
        });
        return;
      }

      var el = node.closest("a");
      if (!el) return;
      var href = el.getAttribute("href") || "";
      var text = (el.textContent || "").replace(/\s+/g, " ").trim().slice(0, 60);

      if (href.indexOf("github.com") !== -1) {
        track("outbound_click", {
          destination: "github.com",
          href: href,
          text: text,
          page: window.location.pathname,
        });
      } else if (href.indexOf("runaic.com") !== -1 && href.indexOf("docs.runaic.com") === -1) {
        track("outbound_click", {
          destination: "runaic.com",
          href: href,
          text: text,
          page: window.location.pathname,
        });
      }
    },
    true
  );
})();
