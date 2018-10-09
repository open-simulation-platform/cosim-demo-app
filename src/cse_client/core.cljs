(ns cse-client.core
  (:require [kee-frame.core :as k]
            [re-frame.core :as rf]
            [cse-client.view :as view]
            [cse-client.controller :as controller]))

(enable-console-print!)

(def routes
  [["/" :index]
   ["/modules/:module" :module]
   ["/trend/:module/:signal" :trend]])

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