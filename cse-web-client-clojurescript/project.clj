(defproject cse-web-client-clojurescript "1.0.0"
  :min-lein-version "2.0.0"
  :dependencies [[kee-frame "0.2.8-SNAPSHOT"]
                 [org.clojure/clojurescript "1.10.312"]
                 [org.clojure/clojure "1.9.0"]
                 [expound "0.7.1"]
                 [ring/ring-core "1.7.0"]]

  :resource-paths ["resources" "target"]
  :clean-targets ^{:protect false} ["target/public"]

  :profiles {:dev {:dependencies [[com.bhauman/figwheel-main "0.1.9"]
                                  [com.bhauman/rebel-readline-cljs "0.1.4"]
                                  [binaryage/devtools "0.9.10"]
                                  [day8.re-frame/re-frame-10x "0.3.3-react16"]]}}
  :aliases {"fig"       ["trampoline" "run" "-m" "figwheel.main"]
            "fig:build" ["trampoline" "run" "-m" "figwheel.main" "-b" "dev" "-r"]})
