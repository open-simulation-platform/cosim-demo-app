(ns cse-client.trend
  (:require [cse-client.controller :as controller]
            [reagent.core :as r]
            [cljsjs.plotly]
            [re-frame.core :as rf]
            [cljs.spec.alpha :as s]
            [clojure.string :as str]))

(defn update-chart-data [dom-node trend-values]
  (s/assert ::trend-values trend-values)

  (doseq [{:keys [labels values]} trend-values]
    (js/Plotly.update dom-node (clj->js {:x [labels]
                                         :y [values]}))))

(defn relayout-callback [js-event]
  (let [event (js->clj js-event)
        begin (get event "xaxis.range[0]")
        end (get event "xaxis.range[1]")
        auto? (get event "xaxis.autorange")]
    (cond
      auto?
      (rf/dispatch [::controller/trend-zoom-reset])

      (and begin end)
      (rf/dispatch [::controller/trend-zoom begin end]))))

(defn trend-inner []
  (let [update (fn [comp]
                 (let [{:keys [trend-values]} (r/props comp)]
                   (update-chart-data (r/dom-node comp) trend-values)))]
    (r/create-class
      {:component-did-mount  (fn [comp]
                               (js/Plotly.plot (r/dom-node comp) (clj->js [{:x []
                                                                            :y []}])
                                               (clj->js {:title @(rf/subscribe [::trend-title])
                                                         :xaxis {:title "Time [s]"}}))
                               (.on (r/dom-node comp) "plotly_relayout" relayout-callback))
       :component-did-update update
       :reagent-render       (fn []
                               [:div {:style {:flex "1 1 auto"}}])})))

(defn trend-outer []
  (let [trend-values (rf/subscribe [::trend-values])
        trend-millis (rf/subscribe [::trend-millis])]
    (fn []
      [:div.main
       [trend-inner {:trend-values @trend-values
                     :trend-millis @trend-millis}]])))

(rf/reg-sub ::trend-values :trend-values)
(rf/reg-sub ::trend-millis :trend-millis)
(rf/reg-sub ::trend-title (fn [db]
                            (let [{:keys [module signal causality type]} (-> db :trend-values first)]
                              (str/join " - " [module signal causality type]))))

(defn ascending-points? [tuples]
  (= tuples
     (sort-by :x tuples)))

(s/def ::module string?)
(s/def ::signal string?)
(s/def ::trend-point (s/keys :req-un [::x ::y]))
(s/def ::ascending-points ascending-points?)
(s/def ::trend-data (s/and (s/coll-of ::trend-point :kind vector?) ::ascending-points))
(s/def ::trend-value (s/keys :req-un [::module ::signal ::trend-data]))
(s/def ::trend-values (s/coll-of ::trend-value))
