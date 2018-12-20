(ns cse-client.controller
  (:require [kee-frame.core :as k]
            [kee-frame.websocket :as websocket]
            [cse-client.config :refer [socket-url]]
            [re-frame.loggers :as re-frame-log]
            [clojure.string :as str]))

;; Prevent handler overwriting warnings during cljs reload.
(re-frame-log/set-loggers!
  {:warn (fn [& args]
           (when-not (or (re-find #"^re-frame: overwriting" (first args))
                         (re-find #"^Overwriting controller" (first args)))
             (apply js/console.warn args)))})

(k/reg-controller :trend
                  {:params (fn [route]
                             (when (= :trend (-> route :data :name))
                               (:path-params route)))
                   :start  [::trend]
                   :stop   [::untrend]})

(k/reg-controller :module
                  {:params (fn [route]
                             (when (-> route :data :name (= :module))
                               (-> route :path-params :module)))
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
                        (assoc-in [:trend-values 0] {:labels TrendTimestamps
                                                     :values TrendValues})
                        (update :state merge message)))))


(defn ws-request [db config]
  (merge
    {:modules-loaded?     (boolean (:modules db))
     :connections-loaded? (boolean (:connections db))}
    config))

(k/reg-event-db ::causality-enter
                (fn [db [causality]]
                  (assoc db :active-causality causality)))

(defn socket-command [db cmd]
  {:dispatch [::websocket/send socket-url (ws-request db {:command cmd})]})

(k/reg-event-fx ::module-enter
                (fn [{:keys [db]} [module]]
                  (socket-command db ["module" module])))

(k/reg-event-fx ::module-leave
                (fn [{:keys [db]} _]
                  (socket-command db ["module" nil])))

(k/reg-event-fx ::load
                (fn [{:keys [db]} [folder log-folder]]
                  (socket-command db ["load" folder (or log-folder "")])))

(k/reg-event-fx ::teardown
                (fn [{:keys [db]} _]
                  (socket-command db ["teardown"])))

(k/reg-event-fx ::play
                (fn [{:keys [db]} _]
                  (socket-command db ["play"])))

(k/reg-event-fx ::pause
                (fn [{:keys [db]} _]
                  (socket-command db ["pause"])))

(k/reg-event-fx ::enable-realtime
                (fn [{:keys [db]} _]
                  (socket-command db ["enable-realtime"])))

(k/reg-event-fx ::disable-realtime
                (fn [{:keys [db]} _]
                  (socket-command db ["disable-realtime"])))

(k/reg-event-fx ::untrend
                (fn [{:keys [db]} _]
                  (merge (socket-command db ["untrend"])
                         {:db (assoc db :trend-values [])})))

(k/reg-event-fx ::trend
                (fn [{:keys [db]} [{:keys [module signal causality type]}]]
                  (merge (socket-command db ["trend" module signal causality type])
                         {:db (assoc db :trend-values [{:trend-data []
                                                        :module     module
                                                        :signal     signal
                                                        :causality  causality
                                                        :type       type}]
                                        :trend-title (str/join " - " [module signal causality type]))})))

(k/reg-event-fx ::set-value
                (fn [{:keys [db]} [module signal causality type value]]
                  (socket-command db ["set-value" module signal causality type (str value)])))

(k/reg-event-fx ::trend-zoom
                (fn [{:keys [db]} [begin end]]
                  (socket-command db ["trend-zoom" (str begin) (str end)])))

(k/reg-event-fx ::trend-zoom-reset
                (fn [{:keys [db]} _]
                  (socket-command db ["trend-zoom-reset" (-> db :trend-range str)])))

(k/reg-event-fx ::trend-range
                (fn [{:keys [db]} [new-range]]
                  {:db       (assoc db :trend-range new-range)
                   :dispatch [::trend-zoom-reset]}))
