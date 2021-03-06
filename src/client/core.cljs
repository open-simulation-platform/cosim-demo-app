;; This Source Code Form is subject to the terms of the Mozilla Public
;; License, v. 2.0. If a copy of the MPL was not distributed with this
;; file, You can obtain one at https://mozilla.org/MPL/2.0/.

(ns client.core
  (:require [kee-frame.core :as k]
            [re-frame.core :as rf]
            [client.view :as view]
            [client.controller :as controller]
            [client.config :refer [socket-url]]
            [cljs.reader :as reader]
            [client.localstorage :as storage]
            [client.config :refer [socket-url]]
            [clojure.string :as str]))

(enable-console-print!)

(goog-define debug false)

(def routes
  [["/" :index]
   ["/modules/:module/:causality" :module]
   ["/trend/:index" :trend]
   ["/scenarios" :scenarios]
   ["/scenarios/:id" :scenario]])

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

(defn scenario-percent [db]
  (let [simulation-time     (simulation-time db)
        scenario-start-time (:scenario-start-time db)
        scenario-end-time   (:scenario-end-time db)]
    (when (and scenario-start-time scenario-end-time)
      (-> (/ (- simulation-time scenario-start-time) scenario-end-time) (* 100) (.toFixed 2) (min 100)))))

(defn real-time-factor [db]
  (str
    (some-> db :state :rollingAverageRealTimeFactor (.toFixed 2))
    "/"
    (some-> db :state :totalAverageRealTimeFactor (.toFixed 2))))

(defn real-time-factor-target [db]
  (some-> db :state :realTimeFactorTarget (.toFixed 3)))

(rf/reg-sub :real-time-factor #(real-time-factor %))

(defn real-time? [db]
  (if (some-> db :state :isRealTime)
    "true"
    "false"))

(defn status-data [db]
  (array-map "Algorithm type" "Fixed step"
             "Simulation time" (simulation-time db)
             "Real time factor (total average)" (some-> db :state :totalAverageRealTimeFactor (.toFixed 3))
             "Real time factor (rolling average)" (some-> db :state :rollingAverageRealTimeFactor (.toFixed 3))
             "Number of steps for rolling average" (some-> db :state :stepsToMonitor)
             "Real time factor target" (real-time-factor-target db)
             "Real time target" (real-time? db)
             "Connection status" (get-in db [:kee-frame.websocket/sockets socket-url :state])
             "Path to loaded config folder" (-> db :state :configDir)))

(rf/reg-sub :overview status-data)
(rf/reg-sub :time simulation-time)
(rf/reg-sub :loading? (comp :loading :state))
(rf/reg-sub :loaded? (comp :loaded :state))
(rf/reg-sub :loaded-dir (fn [db]
                          (-> db :state :configDir)))
(rf/reg-sub :log-dir (fn [db]
                       (:log-dir db)))
(rf/reg-sub :prev-paths (fn [db]
                          (:prev-paths db)))
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

(rf/reg-sub :current-module-index #(-> % :current-module-meta :index))

(rf/reg-sub :scenario-percent scenario-percent)

(rf/reg-sub :pages
            (fn [db]
              (let [page-count (:page-count db)
                    all-pages  (range 1 (inc page-count))]
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
                          (map-indexed (fn [idx {:keys [trend-values] :as trend}]
                                         (-> (select-keys trend [:id :label :plot-type])
                                             (assoc :index idx)
                                             (assoc :count (count trend-values))))
                                       (-> db :state :trends))))

(rf/reg-sub :current-page #(:page %))
(rf/reg-sub :vars-per-page #(:vars-per-page %))

(rf/reg-sub :feedback-message #(:feedback-message %))

(rf/reg-sub :show-success-feedback-messages :show-success-feedback-messages)

(defn validate-event [db event]
  (let [module-tree     (-> db :state :module-data :fmus)
        model-valid?    (->> module-tree (map :name) (filter #(= (:model event) %)) seq boolean)
        variable-valid? (some->> module-tree
                                 (filter #(= (:model event) (:name %)))
                                 first
                                 :variables
                                 (filter (fn [{:keys [name]}]
                                           (= name (:variable event))))
                                 seq
                                 boolean)]
    (-> event
        (assoc :model-valid? model-valid? :variable-valid? variable-valid? :valid? (and model-valid? variable-valid?))
        (assoc :validation-message (cond
                                     (not model-valid?) "Can't find a model with this name"
                                     (not variable-valid?) "Can't find a variable with this name")))))

(defn merge-defaults [db scenario]
  (let [new-events (->> scenario
                        :events
                        (mapv (fn [event]
                                (merge (:defaults scenario) event)))
                        (mapv (fn [event]
                                (validate-event db event)))
                        (sort-by :time))]
    (-> scenario
        (assoc :events new-events)
        (assoc :valid? (every? :valid? new-events)))))

(rf/reg-sub :scenarios (fn [db]
                         (let [running (-> db :state :running-scenario)]
                           (->> (get-in db [:state :scenarios])
                                (map (fn [filename]
                                       {:id       filename
                                        :running? (= filename running)}))))))

(rf/reg-sub :scenario-running?
            (fn [db [_ id]]
              (-> db
                  :state
                  :running-scenario
                  (= id))))

(rf/reg-sub :any-scenario-running?
            (fn [db]
              (-> db :state :running-scenario seq)))


(rf/reg-sub :scenario-start-time
            (fn [db]
              (:scenario-start-time db)))

(rf/reg-sub :scenario-current-time
            (fn [db]
              (let [scenario-start-time (:scenario-start-time db)
                    simulation-time     (-> db :state :time)]
                (- simulation-time scenario-start-time))))

(rf/reg-sub :scenario (fn [db]
                        (->> db
                             :state
                             :scenario
                             (merge-defaults db))))

(rf/reg-sub :get-key
            (fn [db [_ key]]
              (key db)))

(rf/reg-sub :get-state-key
            (fn [db [_ key]]
              (-> db :state key)))

(rf/reg-sub :scenario-id #(:scenario-id %))

(rf/reg-sub :has-manipulator?
            (fn [db [_ index gui-type value-reference]]
              (let [manip-vars (-> db :state :manipulatedVariables)]
                (seq (filter (fn [{:keys [slaveIndex type valueReference]}]
                               (and (= index slaveIndex)
                                    (= gui-type type)
                                    (= value-reference valueReference))) manip-vars)))))

(rf/reg-sub :error (fn [db]
                     (when (-> db :state :executionState (= "COSIM_EXECUTION_ERROR"))
                       {:last-error-code    (-> db :state :lastErrorCode)
                        :last-error-message (-> db :state :lastErrorMessage)})))

(rf/reg-sub :error-dismissed #(:error-dismissed %))

(rf/reg-sub :plot-height #(:plot-height %))

(rf/reg-sub :config-dir (comp :configDir :state))

(rf/reg-sub :plot-config-changed? #(:plot-config-changed? %))

(rf/reg-sub :lib-version (fn [db]
                           (-> db :state :libVersion)))

(k/start! {:routes         routes
           :hash-routing?  true
           :debug?         (if debug {:blacklist #{::controller/socket-message-received}} false)
           :root-component [view/root-comp]
           :initial-db     {:page                           1
                            :vars-per-page                  20
                            :prev-paths                     (reader/read-string (storage/get-item "cosim-paths"))
                            :show-success-feedback-messages (reader/read-string (storage/get-item "show-success-feedback-message"))
                            :plot-config-changed?           false}})
