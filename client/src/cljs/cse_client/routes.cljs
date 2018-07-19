(ns cse-client.routes)

(def routes ["/" {""     :index
                  "sub1" {""      :sub1
                          "/sub2" :sub2}}])