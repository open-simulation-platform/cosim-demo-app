(ns cse-client.routes)

(def routes ["/" {""     :index
                  "sub1" {""      :sub1
                          "/rest" :rest-demo}}])