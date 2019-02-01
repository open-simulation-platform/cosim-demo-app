(ns cse-client.trend
  (:require [cse-client.controller :as controller]
            [reagent.core :as r]
            [cljsjs.plotly]
            [re-frame.core :as rf]
            [cljs.spec.alpha :as s]
            [clojure.string :as str]))

(def range-configs
  [{:seconds 10
    :text    "10s"}
   {:seconds 30
    :text    "30s"}
   {:seconds 60
    :text    "1m"}
   {:seconds (* 60 5)
    :text    "5m"}
   {:seconds (* 60 10)
    :text    "10m"}
   {:seconds (* 60 20)
    :text    "20m"}])

(defn range-selector [trend-range {:keys [text seconds]}]
  ^{:key text}
  [:button.ui.button
   {:on-click #(rf/dispatch [::controller/trend-range seconds])
    :class    (if (= trend-range seconds) "active" "")}
   text])

(defn trend-title [{:keys [module signal causality type]}]
  (str/join " - " [module signal causality type]))

(defn new-series [trend-variable]
  {:name (trend-title trend-variable)
   :x    []
   :y    []})

(defn maybe-update-series [dom-node trend-values]
  (let [num-series (-> dom-node .-data .-length)]
    (when (not= num-series (count trend-values))
      (doseq [_ (range num-series)]
        (js/Plotly.deleteTraces dom-node 0))
      (doseq [trend-variable trend-values]
        (js/Plotly.addTraces dom-node (clj->js (new-series trend-variable)))))))


(defn update-chart-data [dom-node trend-values]
  (s/assert ::trend-values trend-values)
  (let [init-data {:x [] :y []}
        data (reduce (fn [data {:keys [labels values]}]
                       (-> data
                           (update :x conj labels)
                           (update :y conj values)))
                     init-data trend-values)]
    (maybe-update-series dom-node trend-values)
    (js/Plotly.update dom-node (clj->js data))))

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
                                               (clj->js {:xaxis              {:title "Time [s]"}
                                                         :autosize           true
                                                         :use-resize-handler true}))
                               (.on (r/dom-node comp) "plotly_relayout" relayout-callback))
       :component-did-update update
       :reagent-render       (fn []
                               [:div.column])})))

(defn trend-outer []
  (let [trend-values (rf/subscribe [::trend-values])
        trend-range (rf/subscribe [::trend-range])]
    (fn []
      [:div.ui.one.column.grid
       [:div.two.column.row
        [:div.column
         (doall (map (partial range-selector @trend-range) range-configs))]
        [:div.column
         [:button.ui.button.right.floated {:on-click #(rf/dispatch [::controller/untrend])} "Remove all"]]]
       [:div.one.column.row
        [trend-inner {:trend-values @trend-values}]]])))

(rf/reg-sub ::trend-values #(-> % :state :trend-values))
(rf/reg-sub ::trend-range :trend-range)

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
