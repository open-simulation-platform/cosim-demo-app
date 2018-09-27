(ns cse-client.trend
  (:require [reagent.core :as r]
            [cljsjs.highstock]
            [re-frame.core :as rf]
            [cljs.spec.alpha :as s]
            [kee-frame.core :as k]))

(def default-series
  {:animation false
   :data      []
   :marker    {:enabled false}
   :name      "Module and signal info here"})

(defn new-series [{:keys [module signal]}]
  (assoc default-series :name (str module ": " signal)))

(defn maybe-update-series [chart trend-values]
  (let [num-series (-> chart .-series count)]               ;; Use (- 1) when stockChart with navigator is enabled
    (when (not= num-series (count trend-values))
      (doseq [index (range num-series)]
        (-> chart
            .-series
            (aget 0)
            (.remove false)))
      (doseq [trend-variable trend-values]
        (-> chart
            (.addSeries (clj->js (new-series trend-variable)) false))))))

(defn indexed [s]
  (map vector (iterate inc 0) s))

(defn update-chart-data [chart trend-values]
  (s/assert ::trend-values trend-values)
  (maybe-update-series chart trend-values)
  (doseq [[idx {:keys [trend-data]}] (indexed trend-values)]
    (-> chart
        .-series
        (aget idx)
        (.setData (clj->js trend-data) false)))
  (.redraw chart false))

(def range-configs
  [{:millis (* 1000 10)
    :type   "second"
    :text   "10s"}
   {:millis (* 1000 30)
    :type   "second"
    :text   "30s"}
   {:millis (* 1000 60)
    :type   "minute"
    :text   "1m"}
   {:millis (* 1000 60 5)
    :type   "minute"
    :text   "5m"}
   {:millis (* 1000 60 10)
    :type   "minute"
    :text   "10m"}
   {:millis (* 1000 60 20)
    :type   "minute"
    :text   "20m"}])

(defn on-selection [event]
  (if-let [x-axis (some-> event (aget "xAxis") (aget 0))]
    (rf/dispatch [:cybersea.controller.trend/update-zoom {:min (-> x-axis .-min)
                                                          :max (-> x-axis .-max)}])
    (rf/dispatch [:cybersea.controller.trend/update-zoom nil])))

(def default-config
  {:chart     {:type     "line"
               :zoomType "xy"
               :events   {:selection on-selection}}
   :xAxis     {:type                 "datetime"
               :dateTimeLabelFormats {:millisecond "%H:%M:%S.%L"}
               :gridLineWidth        1}
   :yAxis     {:opposite true
               :title    {:text nil}}
   :animation false
   :series    default-series
   :title     {:text ""}
   :legend    {:enabled true}})

(defn range-selector [trend-millis {:keys [text millis]}]
  ^{:key text}
  [:button.btn.btn-default {:on-click #(rf/dispatch [:cybersea.controller.trend/update-millis millis]) :class (if (= trend-millis millis) "selected" "")} text])

(defn trend-inner []
  (let [chart (atom nil)
        update (fn [comp]
                 (let [{:keys [trend-values]} (r/props comp)]
                   #_(when-not (s/valid? ::trend-values trend-values)
                       (throw (ex-info "Invalid trend data, excerpt here " {:trend-values trend-values})))
                   (update-chart-data @chart trend-values)))]
    (r/create-class
      {:component-did-mount    (fn [comp]
                                 (reset! chart (.chart js/Highcharts "charty" (clj->js default-config)))
                                 (update comp))
       :component-will-unmount #(some-> @chart .destroy)
       :component-did-update   update
       :reagent-render         (fn [comp]
                                 [:div {:style {:flex "1 1 auto"}}
                                  (doall (map (partial range-selector (:trend-millis comp)) range-configs))
                                  [:div#charty]])})))

(defn trend-outer []
  (let [trend-values (rf/subscribe [::trend-values])
        trend-millis (rf/subscribe [::trend-millis])]
    (fn []
      [:div.main
       [trend-inner {:config       default-config
                     :trend-values @trend-values
                     :trend-millis @trend-millis}]])))

(rf/reg-sub ::trend-values :trend-values)
(rf/reg-sub ::trend-millis :trend-millis)

(defn ascending-tuples? [tuples]
  (= tuples
     (sort-by first tuples)))

(s/def ::module string?)
(s/def ::signal string?)
(s/def ::trend-tuple (s/tuple number? number?))
(s/def ::ascending-tuples ascending-tuples?)
(s/def ::trend-data (s/and (s/coll-of ::trend-tuple :kind vector?) ::ascending-tuples))
(s/def ::trend-value (s/keys :req-un [::module ::signal ::trend-data]))
(s/def ::trend-values (s/coll-of ::trend-value))