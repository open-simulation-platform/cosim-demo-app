(ns cse-client.core
  (:require [kee-frame.core :as k]
            [re-frame.core :as rf]
            [cse-client.view :as view]
            [cse-client.controller :as controller]
            [cse-client.config :refer [socket-url]]))

(enable-console-print!)

(def routes
  [["/" :index]
   ["/modules/:module" :module]
   ["/trend/:module/:signal/:causality/:type" :trend]])

(defn causalities [db]
  (some->> db
           :state
           :module
           :signals
           (map :causality)
           distinct))

(defn active-causality [db]
  (or ((set (causalities db)) (:active-causality db))
      (-> db
          causalities
          first)))

(defn simulation-time [db]
  (some-> db :state :time (.toFixed 3)))

(defn status-data [db]
  {
   "Algorithm type"               "Fixed step"
   "Simulation time"              (simulation-time db)
   "Real time factor"             "0.9998324"
   "Step size"                    "0.1 s"
   "Connection status"            (get-in db [:kee-frame.websocket/sockets socket-url :state])
   "CPU load"                     "12 %"
   "Total memory"                 "121 MB"
   "Path to loaded config folder" (-> db :state :configDir)})

(rf/reg-sub :overview status-data)
(rf/reg-sub :time simulation-time)
(rf/reg-sub :loaded? (comp :loaded :state))
(rf/reg-sub :status (comp :status :state))
(rf/reg-sub :module (comp :module :state))
(rf/reg-sub :modules (comp :modules :state))
(rf/reg-sub :causalities causalities)
(rf/reg-sub :active-causality active-causality)
(rf/reg-sub :signals (fn [db]
                       (some->> db
                                :state
                                :module
                                :signals
                                (filter (fn [signal]
                                          (= (active-causality db)
                                             (:causality signal))))
                                (sort-by :name))))

(k/start! {:routes         routes
           :hash-routing?  true
           :debug?         {:blacklist #{::controller/socket-message-received}}
           :root-component [view/root-comp]
           :initial-db     {:trend-values []}})