(ns cse-client.core
  (:require [kee-frame.core :as k]
            [kee-frame.websocket :as websocket]
            [re-frame.core :as rf]))

(def socket-url "ws://localhost:8000/ws")

(enable-console-print!)

(def routes ["/" :index])

(k/reg-controller :websocket-controller
                  {:params #(when (-> % :handler (= :index)) true)
                   :start  [:start-websockets]
                   :stop   [:stop-websockets]})

(k/reg-event-fx :start-websockets
                (fn [_ _]
                  {::websocket/open {:path         socket-url
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
                (fn [_ _]
                  {:dispatch [::websocket/send socket-url (ws-request "play")]}))


(k/reg-event-fx :pause
                (fn [_ _]
                  {:dispatch [::websocket/send socket-url (ws-request "pause")]}))

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