(ns cse-client.core
  (:require [kee-frame.core :as k]
            [ajax.core :as ajax]
            [re-interval.core :as re-interval]))

(re-interval/register-interval-handlers :poll nil 1000)


(enable-console-print!)

(def routes ["/" {""     :index
                  "sub1" {""      :sub1
                          "/rest" :rest-demo}}])

(k/reg-controller :state-poll-controller
                  {:params #(when (-> % :handler (= :index)) true)
                   :start  [:poll/start]
                   :stop   [:poll/stop]})

(k/reg-chain :poll/tick
             (fn [_ _]
               {:http-xhrio {:method          :get
                             :uri             "/rest-test"
                             :response-format (ajax/json-response-format)}})
             (fn [{:keys [db]} [_ state]]
               {:db (assoc db :state state)}))

(defn root-comp []
  [:div
   [:ul
    [:li [:a {:href (k/path-for [:index])} "Index"]]
    [:li [:a {:href (k/path-for [:sub1])} "sub1"]]
    [:li [:a {:href (k/path-for [:rest-demo])} "This one is real and will load the REST"]]]
   [:h3 "You navigated to:"]
   [k/switch-route :handler
    :index "This is INDEX!!"
    :sub1 "SUB1 pagey"
    :rest-demo "You will now get an alert with downloaded simulator status"
    nil [:div "Loading..."]]])

(k/start! {:routes         routes
           :hash-routing?  true
           :debug?         true
           :root-component [root-comp]
           :initial-db     {}})