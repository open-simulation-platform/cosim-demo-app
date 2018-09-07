(ns cse-client.core
  (:require [kee-frame.core :as k]
            [kee-frame.websocket :as websocket]
            [ajax.core :as ajax]
            [re-interval.core :as re-interval]
            [re-frame.core :as rf]))

(re-interval/register-interval-handlers :poll nil 1000)


(enable-console-print!)

(def routes ["/" :index])

(k/reg-controller :websocket-controller
                  {:params #(when (-> % :handler (= :index)) true)
                   :start  [:start-websockets]
                   :stop   [:stop-websockets]})

(k/reg-event-fx :start-websockets
                (fn [_ _]
                  {::websocket/open {:path         "/ws"
                                     :dispatch     ::socket-message-received
                                     :format       :json-kw
                                     :wrap-message identity}}))

(k/reg-event-db ::socket-message-received
                (fn [db [{message :message}]]
                  (assoc db :state message)))

(defn ws-request [command]
  (merge
    (when command
      {:command command})
    {:module      "Clock"
     :modules     false
     :connections false}))

(k/reg-event-fx :play
                (fn [{:keys [db]} _]
                  {:dispatch [::websocket/send "/ws" (ws-request "play")]}))


(k/reg-chain :pause
             (fn [{:keys [db]} _]
               {:dispatch [::websocket/send "/ws" (ws-request "pause")]}))

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