(ns cse-client.core
  (:require [kee-frame.core :as k]
            [kee-frame.websocket :as websocket]
            [re-frame.core :as rf]
            [cse-client.trend :as trend]))

(def socket-url "ws://localhost:8000/ws")

(enable-console-print!)

(def routes
  [["/" :index]
   ["/modules/:name" :module]
   ["/trend/:module/:signal" :trend]])

(k/reg-controller :websocket-controller
                  {:params (constantly true)
                   :start  [:start-websockets]})

(k/reg-event-fx :start-websockets
                (fn [_ _]
                  {::websocket/open {:path         socket-url
                                     :dispatch     ::socket-message-received
                                     :format       :json-kw
                                     :wrap-message identity}}))

(k/reg-event-db ::socket-message-received
                (fn [db [{message :message}]]
                  (let [value (-> message :module :signals first :value)]
                    (-> db
                        (update-in [:trend-values 0 :trend-data] conj [value value])
                        (update :state merge message)))))

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

(k/reg-event-fx :trend
                (fn [{:keys [db]} [module signal]]
                  {:db          (assoc db :trend-values [{:trend-data []
                                                          :module     "Clock"
                                                          :signal     "Clock"}])
                   :navigate-to [:trend {:module module :signal signal}]}))

(rf/reg-sub :state :state)

(defn module-listing [signals module-name]
  [:div
   [:a {:href (k/path-for [:index])} "Back to modules"]
   [:ul
    (map (fn [signal]
           [:li {:key (str name "_ " (:name signal))} (:name signal) ": " (:value signal)
            [:button.ui.button {:on-click #(rf/dispatch [:trend module-name (:name signal)])} "Trend"]])
         signals)]])

(defn modules-menu [modules]
  [:div
   [:h3 "Modules"]
   [:ul
    (map (fn [module]
           [:li {:key module} [:a {:href (k/path-for [:module {:name module}])} module]])
         modules)]])

(defn controls []
  [:div.ui.buttons
   [:button.ui.button {:on-click #(rf/dispatch [:play])} "Play"]
   [:button.ui.button {:on-click #(rf/dispatch [:pause])} "Pause"]])

(defn root-comp []
  (let [{:keys [module modules]} @(rf/subscribe [:state])
        {:keys [name signals]} module]
    [:div
     [controls]
     [k/switch-route (comp :name :data)
      :trend [trend/trend-outer]
      :module [module-listing signals name]
      :index [modules-menu modules]
      nil [:div "Loading..."]]]))

(k/start! {:routes         routes
           :hash-routing?  true
           :debug?         {:blacklist #{::socket-message-received}}
           :root-component [root-comp]
           :initial-db     {:trend-values []}})