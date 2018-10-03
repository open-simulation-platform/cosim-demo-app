(ns cse-client.controller
  (:require [kee-frame.core :as k]
            [kee-frame.websocket :as websocket]))

(def socket-url "ws://localhost:8000/ws")

(k/reg-controller :trend
                  {:params (fn [route]
                             (when (= :trend (-> route :data :name))
                               (:path-params route)))
                   :start  [::trend]
                   :stop   [::untrend]})

(k/reg-controller :module
                  {:params (fn [route]
                             (when (= :module (-> route :data :name))
                               (-> route :path-params :name)))
                   :start  [::module-enter]})

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
                  (let [{:keys [TrendTimestamps TrendValues]} (-> message :trendSignals first)]
                    (-> db
                        (assoc-in [:trend-values 0 :trend-data] (map vector TrendTimestamps TrendValues))
                        (update :state merge message)))))


(defn ws-request [db config]
  (merge
    {:modules-loaded?     (boolean (:modules db))
     :connections-loaded? (boolean (:connections db))}
    config))

(k/reg-event-fx ::module-enter
                (fn [{:keys [db]} [module]]
                  {:dispatch [::websocket/send socket-url (ws-request db {:module module})]}))

(k/reg-event-fx ::module-leave
                (fn [{:keys [db]} _]
                  {:dispatch [::websocket/send socket-url (ws-request db {:module nil})]}))

(k/reg-event-fx ::play
                (fn [{:keys [db]} _]
                  {:dispatch [::websocket/send socket-url (ws-request db {:command ["play"]})]}))

(k/reg-event-fx ::pause
                (fn [{:keys [db]} _]
                  {:dispatch [::websocket/send socket-url (ws-request db {:command ["pause"]})]}))

(k/reg-event-fx ::untrend
                (fn [{:keys [db]} _]
                  {:dispatch [::websocket/send socket-url (ws-request db {:command ["untrend"]})]
                   :db       (assoc db :trend-values [])}))

(k/reg-event-fx ::trend
                (fn [{:keys [db]} [{:keys [module signal]}]]
                  {:dispatch [::websocket/send socket-url (ws-request db {:command ["trend" module signal]})]
                   :db       (assoc db :trend-values [{:trend-data []
                                                       :module     module
                                                       :signal     signal}])}))
