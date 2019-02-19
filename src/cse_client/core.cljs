(ns cse-client.core
  (:require [kee-frame.core :as k]
            [re-frame.core :as rf]
            [cse-client.view :as view]
            [cse-client.controller :as controller]
            [cse-client.config :refer [socket-url]]))

(enable-console-print!)

(def routes
  [["/" :index]
   ["/modules/:module/:causality" :module]
   ["/trend/:index" :trend]
   ["/guide" :guide]])

(def sort-order
  (let [order ["input" "independent" "parameter" "calculatedParameter" "local" "internal" "output"]]
    (zipmap order (range (count order)))))

(defn by-order [s1 s2]
  (compare (sort-order s1) (sort-order s2)))

(defn find-module [db module]
  (some->> (get-in db [:state :module-data :fmus])
           (filter #(= module (:name %)))
           first))

(defn causalities [db module]
  (when-let [variables (some-> db (find-module module) :variables)]
    (some->> variables
             (map :causality)
             distinct
             (sort by-order))))

(defn active-causality [db]
  (let [module (:current-module db)]
    (or ((set (causalities db module)) (:active-causality db))
        (-> db
            (causalities module)
            first))))

(defn simulation-time [db]
  (some-> db :state :time (.toFixed 3)))

(defn real-time-factor [db]
  (some-> db :state :realTimeFactor (.toFixed 3)))

(rf/reg-sub :real-time-factor #(real-time-factor %))

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

(rf/reg-sub :module-routes (fn [db]
                             (let [modules (-> db :state :module-data :fmus)]
                               (map (fn [module]
                                      {:name      (:name module)
                                       :causality (first (causalities db (:name module)))})
                                    modules))))

(rf/reg-sub :module-active? (fn [db]
                              (= (:current-module db) (->> db :state :module :name))))

(rf/reg-sub :current-module #(:current-module %))

(rf/reg-sub :pages
            (fn [db]
              (let [page-count (:page-count db)
                    all-pages (range 1 (inc page-count))]
                all-pages)))

(rf/reg-sub :module-signals (fn [db]
                              (:viewing db)))

(rf/reg-sub :causalities (fn [db] (causalities db (:current-module db))))
(rf/reg-sub :active-causality active-causality)
(rf/reg-sub :active-trend-index :active-trend-index)

(rf/reg-sub :signal-value (fn [db [_ module name causality type]]
                            (->> db
                                 :state
                                 :module
                                 :signals
                                 (filter (fn [var]
                                           (and (= (:name var) name)
                                                (= (:causality var) causality)
                                                (= (:type var) type))))
                                 first
                                 :value)))

(rf/reg-sub :trend-info (fn [db _]
                          (map-indexed
                            (fn [idx {:keys [trend-values] :as trend}]
                              (-> (select-keys trend [:id :label :plot-type])
                                  (assoc :index idx)
                                  (assoc :count (count trend-values)))) (-> db :state :trends))))

(rf/reg-sub :active-guide-tab :active-guide-tab)

(rf/reg-sub :current-page #(:page %))
(rf/reg-sub :vars-per-page #(:vars-per-page %))

(rf/reg-sub :feedback-message #(:feedback-message %))

(k/start! {:routes         routes
           :hash-routing?  true
           :debug?         {:blacklist #{::controller/socket-message-received}}
           :root-component [view/root-comp]
           :initial-db     {:trend-range      10
                            :active-guide-tab "About"
                            :page             1
                            :vars-per-page    20}})
