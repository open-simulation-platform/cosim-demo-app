(ns cse-client.core
  (:require [kee-frame.core :as k]
            [ajax.core :as ajax]))

(enable-console-print!)

(def routes ["/" {""     :index
                  "sub1" {""      :sub1
                          "/rest" :rest-demo}}])

(k/reg-controller :rest-demo-controller
                  {:params #(when (-> % :handler (= :rest-demo)) true)
                   :start  [:load-rest-demo]})

(k/reg-chain :load-rest-demo
             (fn [_ _]
               {:http-xhrio {:method          :get
                             :uri             "/rest-test"
                             :response-format (ajax/json-response-format)}})
             (fn [_ [_ rest-of-it]]
               (js/alert (str "GOT ME SOME REST from \"/rest-test\": " rest-of-it))))

(defn root-comp []
  [:div
   [:ul
    [:li [:a {:href (k/path-for [:index])} "Index"]]
    [:li [:a {:href (k/path-for [:sub1])} "sub1"]]
    [:li [:a {:href (k/path-for [:rest-demo])} "This one is real and will load the REST"]]]
   [:h3 "You navigated to:"]
   [k/switch-route :handler
    :index "This is INDEX!!"
    :sub1 "SUB1 page"
    :rest-demo "You will now get an alert with downloaded simulator status"
    nil [:div "Loading..."]]])

(k/start! {:routes         routes
           :hash-routing?  true
           :debug?         true
           :root-component [root-comp]
           :initial-db     {}})