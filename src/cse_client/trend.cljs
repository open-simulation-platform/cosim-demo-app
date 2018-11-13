(ns cse-client.trend
  (:require [reagent.core :as r]
            [cljsjs.plotly]
            [re-frame.core :as rf]
            [cljs.spec.alpha :as s]))

(defn indexed [s]
  (map vector (iterate inc 0) s))

(defn update-chart-data [trend-values]
  (s/assert ::trend-values trend-values)

  (doseq [[idx {:keys [labels values]}] (indexed trend-values)]
    (js/Plotly.extendTraces "charty" (clj->js {:x [labels]
                                               :y [values]})
                            (clj->js [0]))))

(defn trend-inner []
  (let [update (fn [comp]
                 (let [{:keys [trend-values]} (r/props comp)]
                   (update-chart-data trend-values)))]
    (r/create-class
      {:component-did-mount  (fn [comp]
                               (js/Plotly.plot "charty" (clj->js [{:x []
                                                                   :y []}])))
       :component-did-update update
       :reagent-render       (fn [comp]
                               [:div {:style {:flex "1 1 auto"}}
                                [:div#charty]])})))

(defn trend-outer []
  (let [trend-values (rf/subscribe [::trend-values])
        trend-millis (rf/subscribe [::trend-millis])]
    (fn []
      [:div.main
       [trend-inner {:trend-values @trend-values
                     :trend-millis @trend-millis}]])))

(rf/reg-sub ::trend-values :trend-values)
(rf/reg-sub ::trend-millis :trend-millis)

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
