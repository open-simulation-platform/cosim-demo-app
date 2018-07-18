(defproject cse-client "1.0.0"
  :min-lein-version "2.0.0"
  :dependencies [[kee-frame "0.2.4" :scope "provided"]
                 [day8.re-frame/http-fx "0.1.6" :scope "provided"]
                 [cljs-ajax "0.7.3" :scope "provided"]
                 [org.clojure/tools.reader "1.3.0-alpha3"]
                 [cljsjs/bootstrap "3.3.5-0" :scope "provided"]
                 [org.clojure/clojurescript "1.10.339" :scope "provided"]
                 [org.clojure/clojure "1.9.0"]]
  :plugins [[lein-count "1.0.7"]
            [lein-figwheel "0.5.16"]
            [lein-cljsbuild "1.1.7"]]

  :clean-targets ^{:protect false} [:target-path :compile-path "resources/public/js/compiled"]

  :source-paths ["src/clj" "src/cljc"]

  :cljsbuild {:builds [{:id           "app"
                        :source-paths ["src/cljs" "src/cljc"]
                        :figwheel     true
                        :compiler     {:main                 cse-client.core
                                       :asset-path           "/js/compiled/out"
                                       :output-to            "../resources/public/js/compiled/app.js"
                                       :output-dir           "../resources/public/js/compiled/out"
                                       :source-map-timestamp true
                                       :parallel-build       true
                                       :closure-defines      {cse-client.core/debug                 true
                                                              "re_frame.trace.trace_enabled_QMARK_" true}
                                       :preloads             [devtools.preload day8.re-frame-10x.preload]
                                       :external-config      {:devtools/config {:features-to-install [:formatters]}}}}
                       {:id           "min"
                        :source-paths ["src/cljs" "src/cljc"]
                        :compiler     {:output-to      "../resources/public/js/compiled/app.js"
                                       :optimizations  :advanced
                                       :parallel-build true}}]}

  :figwheel {:css-dirs ["resources/public/css"]}

  :profiles {:dev          [:project/dev :profiles/dev]
             :profiles/dev {}
             :project/dev  {:dependencies [[figwheel "0.5.16"]
                                           [figwheel-sidecar "0.5.16"]
                                           [binaryage/devtools "0.9.10"]
                                           [day8.re-frame/re-frame-10x "0.3.3-react16"]]}})
