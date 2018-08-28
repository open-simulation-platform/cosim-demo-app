(ns cse-client.core
  (:require [kee-frame.core :as k]
            [ajax.core :as ajax]
            [re-interval.core :as re-interval]
            [re-frame.core :as rf]))

(re-interval/register-interval-handlers :poll nil 1000)


(enable-console-print!)

(def routes ["/" :index])

(k/reg-controller :state-poll-controller
                  {:params #(when (-> % :handler (= :index)) true)
                   :start  [:poll/start]
                   :stop   [:poll/stop]})

(k/reg-chain :poll/tick
             (fn [_ _]
               {:http-xhrio {:method          :get
                             :uri             "/rest-test"
                             :response-format (ajax/json-response-format {:keywords? true})}})
             (fn [{:keys [db]} [state]]
               {:db (assoc db :state state)}))

(k/reg-chain :play
             (fn [_ _]
               {:http-xhrio {:method          :get
                             :uri             "/play"
                             :response-format (ajax/json-response-format {:keywords? true})}})
             (fn [_ _]))

(k/reg-chain :pause
             (fn [_ _]
               {:http-xhrio {:method          :get
                             :uri             "/pause"
                             :response-format (ajax/json-response-format {:keywords? true})}})
             (fn [_ _]))

(rf/reg-sub :state :state)

(defn root-comp []
  (let [{:keys [name status signalValue]} @(rf/subscribe [:state])]
    [:div
     [:h3 "Simulator:"]
     [:ul
      [:li "Name: " name]
      [:li "Status: " status]
      [:li "Signal value: " signalValue]]
     [:p
      [:button {:on-click #(rf/dispatch [:play])} "Play"]
      [:button {:on-click #(rf/dispatch [:pause])} "Pause"]]]))

(k/start! {:routes         routes
           :hash-routing?  true
           :debug?         true
           :root-component [root-comp]
           :initial-db     {}})