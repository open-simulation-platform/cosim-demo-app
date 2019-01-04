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
   ["/trend" :trend]
   ["/guide" :guide]])

(def sort-order
  (let [order ["input" "independent" "parameter" "calculatedParameter" "local" "internal" "output"]]
    (zipmap order (range (count order)))))

(defn by-order [s1 s2]
  (compare (sort-order s1) (sort-order s2)))

(defn causalities [db]
  (some->> db
           :state
           :module
           :signals
           (map :causality)
           distinct
           (sort by-order)))

(defn active-causality [db]
  (or ((set (causalities db)) (:active-causality db))
      (-> db
          causalities
          first)))

(defn editable? [{:keys [type causality] :as variable}]
  (if (and (#{"input" "parameter"} causality)
           (#{"Real" "Integer"} type))
    (assoc variable :editable? true)
    variable))

(defn simulation-time [db]
  (some-> db :state :time (.toFixed 3)))

(defn real-time-factor [db]
  (some-> db :state :realTimeFactor (.toFixed 3)))

(defn real-time? [db]
  (if (some-> db :state :isRealTime)
    "true"
    "false"))

(defn status-data [db]
  {"Algorithm type"               "Fixed step"
   "Simulation time"              (simulation-time db)
   "Real time factor"             (real-time-factor db)
   "Real time target"             (real-time? db)
   "Connection status"            (get-in db [:kee-frame.websocket/sockets socket-url :state])
   "Path to loaded config folder" (-> db :state :configDir)})

(rf/reg-sub :overview status-data)
(rf/reg-sub :time simulation-time)
(rf/reg-sub :loaded? (comp :loaded :state))
(rf/reg-sub :status (comp :status :state))
(rf/reg-sub :realtime? (comp :isRealTime :state))
(rf/reg-sub :module (comp :module :state))
(rf/reg-sub :modules (comp :modules :state))
(rf/reg-sub :causalities causalities)
(rf/reg-sub :active-causality active-causality)
(rf/reg-sub :active-guide-tab :active-guide-tab)
(rf/reg-sub :signals (fn [db]
                       (some->> db
                                :state
                                :module
                                :signals
                                (filter (fn [signal]
                                          (= (active-causality db)
                                             (:causality signal))))
                                (map editable?)
                                (sort-by :name))))
(rf/reg-sub :trend-count #(-> % :state :trend-values count))

(k/start! {:routes         routes
           :hash-routing?  true
           :debug?         {:blacklist #{::controller/socket-message-received}}
           :root-component [view/root-comp]
           :initial-db     {:trend-range 10
                            :active-guide-tab "About"}})
