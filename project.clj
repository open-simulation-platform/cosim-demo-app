(defproject cse-client "1.0-SNAPSHOT"
  :min-lein-version "2.0.0"
  :dependencies [[kee-frame "0.3.3"]
                 [jarohen/chord "0.8.1"]
                 [com.cognitect/transit-clj "0.8.309"]
                 [com.cognitect/transit-cljs "0.8.256"]
                 [reagent "0.8.1"]
                 [re-frame "0.10.6" :exclusions [reagent]]
                 [org.clojure/clojurescript "1.10.439"]
                 [org.clojure/clojure "1.9.0"]
                 [org.clojure/tools.reader "1.3.0"]
                 [fulcrologic/fulcro "2.8.0"]
                 [fulcrologic/semantic-ui-react-wrappers "2.0.4"]
                 [cljsjs/semantic-ui-react "0.84.0-0"]
                 [binaryage/oops "0.7.0"]]
  :plugins [[lein-figwheel "0.5.18"]
            [figwheel-sidecar "0.5.17"]
            [lein-cljsbuild "1.1.7"]
            [lein-count "1.0.9"]
            [cider/cider-nrepl "0.21.1"]]

  :clean-targets ^{:protect false} [:target-path :compile-path "resources/public/static/js/compiled"]

  :cljsbuild {:builds [{:id           "app"
                        :source-paths ["src"]
                        :figwheel     true
                        :compiler     {:main                 cse-client.core
                                       :asset-path           "/static/js/compiled/out"
                                       :output-to            "resources/public/static/js/compiled/app.js"
                                       :output-dir           "resources/public/static/js/compiled/out"
                                       :source-map-timestamp true
                                       :parallel-build       true
                                       :closure-defines      {cse-client.core/debug                 true
                                                              cse-client.view/default-load-dir      ""
                                                              cse-client.view/default-log-dir       ""
                                                              "re_frame.trace.trace_enabled_QMARK_" true}
                                       :preloads             [devtools.preload day8.re-frame-10x.preload]
                                       :external-config      {:devtools/config {:features-to-install [:formatters]}}
                                       :foreign-libs         [{:file     "resources/public/static/js/plotly.min.js"
                                                               :provides ["cljsjs.plotly"]}]}}
                       {:id           "min"
                        :source-paths ["src"]
                        :compiler     {:output-to      "resources/public/static/js/compiled/app.js"
                                       :optimizations  :advanced
                                       :parallel-build true
                                       :foreign-libs   [{:file     "resources/public/static/js/plotly.min.js"
                                                         :provides ["cljsjs.plotly"]}]}}]}

  :figwheel {:css-dirs ["resources/public/static/css"]}

  :profiles {:dev          [:project/dev :profiles/dev]
             :profiles/dev {}
             :project/dev  {:dependencies [[binaryage/devtools "0.9.10"]
                                           [day8.re-frame/re-frame-10x "0.3.3-react16"]
                                           [cider/piggieback "0.4.0"]
                                           [figwheel-sidecar "0.5.18"]]
                            :repl-options {:nrepl-middelware [cider.piggieback/wrap-cljs-repl]}}})
