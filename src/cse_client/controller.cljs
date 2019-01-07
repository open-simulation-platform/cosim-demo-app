(ns cse-client.controller
  (:require [kee-frame.core :as k]
            [kee-frame.websocket :as websocket]
            [cse-client.config :refer [socket-url]]
            [re-frame.loggers :as re-frame-log]
            [cljs.spec.alpha :as s]
            [clojure.string :as str]))

;; Prevent handler overwriting warnings during cljs reload.
(re-frame-log/set-loggers!
  {:warn (fn [& args]
           (when-not (or (re-find #"^re-frame: overwriting" (first args))
                         (re-find #"^Overwriting controller" (first args)))
             (apply js/console.warn args)))})

(k/reg-controller :module
                  {:params (fn [route]
                             (when (-> route :data :name (= :module))
                               (-> route :path-params)))
                   :start  [::module-enter]
                   :stop   [::module-leave]})

(k/reg-controller :websocket-controller
                  {:params (constantly true)
                   :start  [:start-websockets]})

(defn socket-command [cmd]
  {:dispatch [::websocket/send socket-url {:command cmd}]})

(k/reg-chain :start-websockets
             (fn [_ _]
               {::websocket/open {:path         socket-url
                                  :dispatch     ::socket-message-received
                                  :format       :json-kw
                                  :wrap-message identity}})
             (fn [_ _ _]
               (socket-command ["get-module-data"])))

(s/def ::fmu (s/keys :req-un [::name ::index ::variables]))
(s/def ::fmus (s/coll-of ::fmu))
(s/def ::module-data (s/keys :req-un [::fmus]))

(k/reg-event-db ::socket-message-received
                (fn [db [{message :message}]]
                  (when-let [module-data (:module-data message)]
                    (s/assert ::module-data module-data))
                  (update db :state merge message)))

(k/reg-event-fx ::fetch-module-data
                (fn [_ _]
                  (socket-command ["get-module-data"])))

(defn encode-variable [{:keys [name causality type value-reference]}]
  (str/join "," [name causality type value-reference]))

(defn make-signals [db module causality]
  (->> db
       :state
       :module-data
       :fmus
       (filter #(= module (:name %)))
       first
       :variables
       (filter #(= causality (:causality %)))
       (map encode-variable)))

(k/reg-event-fx ::module-enter
                (fn [{:keys [db]} [{:keys [module causality]}]]
                  (merge
                    {:db (assoc db :current-module module
                                   :active-causality causality)}
                    (socket-command (concat ["signals" module] (make-signals db module causality))))))

(k/reg-event-fx ::module-leave
                (fn [{:keys [db]} _]
                  (merge
                    {:db (dissoc db :current-module)}
                    (socket-command ["signals"]))))

(k/reg-event-fx ::load
                (fn [_ [folder log-folder]]
                  (socket-command ["load" folder (or log-folder "")])))

(k/reg-event-fx ::teardown
                (fn [_ _]
                  (socket-command ["teardown"])))

(k/reg-event-fx ::play
                (fn [_ _]
                  (socket-command ["play"])))

(k/reg-event-fx ::pause
                (fn [_ _]
                  (socket-command ["pause"])))

(k/reg-event-fx ::enable-realtime
                (fn [_ _]
                  (socket-command ["enable-realtime"])))

(k/reg-event-fx ::disable-realtime
                (fn [_ _]
                  (socket-command ["disable-realtime"])))

(k/reg-event-fx ::untrend
                (fn [_ _]
                  (socket-command ["untrend"])))

(k/reg-event-fx ::add-to-trend
                (fn [_ [module signal causality type]]
                  (socket-command ["trend" module signal causality type])))

(k/reg-event-fx ::set-value
                (fn [_ [module signal causality type value]]
                  (socket-command ["set-value" module signal causality type (str value)])))

(k/reg-event-fx ::trend-zoom
                (fn [_ [begin end]]
                  (socket-command ["trend-zoom" (str begin) (str end)])))

(k/reg-event-fx ::trend-zoom-reset
                (fn [{:keys [db]} _]
                  (socket-command ["trend-zoom-reset" (-> db :trend-range str)])))

(k/reg-event-fx ::trend-range
                (fn [{:keys [db]} [new-range]]
                  {:db       (assoc db :trend-range new-range)
                   :dispatch [::trend-zoom-reset]}))
