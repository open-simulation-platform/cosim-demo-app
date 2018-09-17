(ns cse-client.core
  (:require [kee-frame.core :as k]
            [kee-frame.websocket :as websocket]
            [re-frame.core :as rf]
            [soda-ash.core :as sa]))

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
                  (update db :state merge message)))

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
  (let [{:keys [module modules]} @(rf/subscribe [:state])
        {:keys [name signals]} module]
    [:div
     [:h3 "Modules"]
     [:ul
      (map (fn [module]
             [:li {:key module} module])
           modules)]
     [:h3 "Selected module: " name]
     [:ul
      (map (fn [{:keys [name value]}]
             [:li {:key (str module "_ " name)} "Signal name: " name
              [:ul
               [:li "Signal value: " value]]])
           signals)]
     [:div.ui.buttons
      [:button.ui.button {:on-click #(rf/dispatch [:play])} "Play"]
      [:button.ui.button {:on-click #(rf/dispatch [:pause])} "Pause"]]]))

(k/start! {:routes         routes
           :hash-routing?  true
           :debug?         {:blacklist #{::socket-message-received}}
           :root-component [root-comp]
           :initial-db     {}})