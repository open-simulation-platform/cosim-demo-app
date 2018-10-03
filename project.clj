(defproject cse-client "1.0-SNAPSHOT"
  :min-lein-version "2.0.0"
  :dependencies [[kee-frame "0.2.8-SNAPSHOT"]
                 [reagent "0.8.1"]
                 [re-frame "0.10.6" :exclusions [reagent]]
                 [cljsjs/highstock "6.0.7-0"]
                 [org.clojure/clojurescript "1.10.339"]
                 [org.clojure/clojure "1.9.0"]]
  :plugins [[lein-figwheel "0.5.16"]
            [lein-cljsbuild "1.1.7"]]

  :clean-targets ^{:protect false} [:target-path :compile-path "resources/public/js/compiled"]

  :cljsbuild {:builds [{:id           "app"
                        :source-paths ["src"]
                        :figwheel     true
                        :compiler     {:main                 cse-client.core
                                       :asset-path           "/js/compiled/out"
                                       :output-to            "resources/public/js/compiled/app.js"
                                       :output-dir           "resources/public/js/compiled/out"
                                       :source-map-timestamp true
                                       :parallel-build       true
                                       :closure-defines      {cse-client.core/debug                 true
                                                              "re_frame.trace.trace_enabled_QMARK_" true}
                                       :preloads             [devtools.preload day8.re-frame-10x.preload]
                                       :external-config      {:devtools/config {:features-to-install [:formatters]}}}}
                       {:id           "min"
                        :source-paths ["src"]
                        :compiler     {:output-to      "resources/public/js/compiled/app.js"
                                       :optimizations  :advanced
                                       :parallel-build true}}]}

  :figwheel {:css-dirs ["resources/public/css"]}

  :profiles {:dev          [:project/dev :profiles/dev]
             :profiles/dev {}
             :project/dev  {:dependencies [[binaryage/devtools "0.9.10"]
                                           [day8.re-frame/re-frame-10x "0.3.3-react16"]]}})
