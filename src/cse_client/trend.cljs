(ns cse-client.trend
  (:require [cse-client.controller :as controller]
            [reagent.core :as r]
            [cljsjs.plotly]
            [re-frame.core :as rf]
            [cljs.spec.alpha :as s]
            [cse-client.components :as c]
            [clojure.string :as str]))

(def id-store (atom nil))

(def plot-container-height "70vh")

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

(def common-layout
  {:autosize           true
   :use-resize-handler true
   :showlegend         true
   :uirevision         true
   :margin             {:l 40 :r 40 :t 20 :pad 5}
   :legend             {:orientation "h"}})

(def trend-layout
  (merge common-layout
         {:xaxis {:title "Time [s]"}}))

(def scatter-layout
  (merge common-layout
         {:xaxis {:autorange true
                  :autotick  true
                  :ticks     ""}}))

(defn- layout-selector [plot-type]
  (case plot-type
    "trend" trend-layout
    "scatter" scatter-layout
    {}))

(defn- namespaced
  "Takes a map and a symbol name or string and creates a new map with namespaced keys as defined by the symbol.
  E.g. (namespaced {:a 1} 'my-ns) -> {:my-ns/a 1}"
  [m ns]
  (into {}
        (map (fn [[k v]] [(keyword (str ns "/" (name k))) v])
             (seq m))))

(def first-signal-ns 'first)
(def second-signal-ns 'second)

(defn- format-data-for-plotting
  "Data for time series plots (trend) are returned as is.
  For XY plots (scatter) pairs of trend-values are merged together to form a plot with x and y values.
  The metadata fields are given namespaces to avoid loosing information when merging the pairs of values."
  [plot-type trend-values]
  (case plot-type
    "trend" trend-values
    "scatter" (map (fn [[a b]]
                     (merge
                       (select-keys a [:xvals :yvals])
                       (select-keys b [:xvals :yvals])
                       (namespaced (dissoc a :xvals :yvals) first-signal-ns)
                       (namespaced (dissoc b :xvals :yvals) second-signal-ns)))
                   (partition 2 trend-values))
    []))

(defn- range-selector [trend-range {:keys [text seconds]}]
  ^{:key text}
  [:button.ui.button
   {:on-click #(rf/dispatch [::controller/trend-range seconds])
    :class    (if (= trend-range seconds) "active" "")}
   text])

(defn plot-type-from-label [label]
  "Expects label to be a string on format 'Time series #a9123ddc-..'"
  (str/trim (first (str/split label "#"))))

(defn- time-series-legend-name [{:keys [module signal causality type]}]
  (str/join " - " [module signal causality type]))

(defn- xy-plot-legend-name [plot]
  (let [first-signal  ((keyword (str first-signal-ns "/" 'signal)) plot)
        second-signal ((keyword (str second-signal-ns "/" 'signal)) plot)]
    (str/join " / " [first-signal second-signal])))

(defn- delete-series [dom-node]
  (let [num-series (-> dom-node .-data .-length)]
    (doseq [_ (range num-series)]
      (js/Plotly.deleteTraces dom-node 0))))

(defn- add-traces [dom-node plots legend-fn]
  (doseq [plot plots]
    (js/Plotly.addTraces dom-node (clj->js {:name (legend-fn plot) :x [] :y []}))))

(defn- maybe-update-series [dom-node trend-values]
  (let [num-series (-> dom-node .-data .-length)]
    (when (not= num-series (count trend-values))
      (doseq [_ (range num-series)]
        (js/Plotly.deleteTraces dom-node 0))
      (case (:plot-type @(rf/subscribe [::active-trend]))
        "trend" (add-traces dom-node trend-values time-series-legend-name)
        "scatter" (add-traces dom-node trend-values xy-plot-legend-name)))))

(defn- update-chart-data [dom-node trend-values trend-id]
  (when-not (= trend-id @id-store)
    (reset! id-store trend-id)
    (delete-series dom-node))
  (s/assert ::trend-values trend-values)
  (let [init-data {:x [] :y []}
        data      (reduce (fn [data {:keys [xvals yvals]}]
                            (-> data
                                (update :x conj xvals)
                                (update :y conj yvals)))
                          init-data trend-values)]
    (maybe-update-series dom-node trend-values)
    (js/Plotly.update dom-node (clj->js data))))

(defn- relayout-callback [js-event]
  (let [event        (js->clj js-event)
        begin        (get event "xaxis.range[0]")
        end          (get event "xaxis.range[1]")
        auto?        (get event "xaxis.autorange")
        active-trend @(rf/subscribe [::active-trend])]
    (when (= (:plot-type active-trend) "trend")
      (cond
        auto?
        (rf/dispatch [::controller/trend-zoom-reset])
        (and begin end)
        (rf/dispatch [::controller/trend-zoom begin end])))))

(defn- set-dom-element-height! [dom-node height]
  (-> dom-node .-style .-height (set! height)))

(defn- trend-inner []
  (let [update-plot (fn [comp]
                      (let [{:keys [trend-values trend-id]} (r/props comp)]
                        (update-chart-data (r/dom-node comp) trend-values trend-id)))
        render-plot (fn [comp]
                      (let [{:keys [plot-type]} (r/props comp)
                            dom-node  (r/dom-node comp)
                            _         (set-dom-element-height! dom-node plot-container-height)
                            layout    (layout-selector plot-type)]
                        (js/Plotly.newPlot dom-node
                                           (clj->js [{:x    []
                                                      :y    []
                                                      :mode "lines"
                                                      :type "scatter"}])
                                           (clj->js layout)
                                           (clj->js {:responsive true}))))]
    (r/create-class
      {:component-did-mount  (fn [comp]
                               (render-plot comp)
                               (.on (r/dom-node comp) "plotly_relayout" relayout-callback))
       :component-did-update update-plot
       :reagent-render       (fn []
                               [:div.column])})))

(defn trend-outer []
  (let [trend-range        (rf/subscribe [::trend-range])
        active-trend       (rf/subscribe [::active-trend])
        active-trend-index (rf/subscribe [:active-trend-index])]
    (fn []
      (let [{:keys [id plot-type label trend-values]} @active-trend
            active-trend-index (int @active-trend-index)
            name               (plot-type-from-label label)]
        [:div.ui.one.column.grid
         [c/text-editor name [::controller/set-label] "Rename plot"]
         [:div.two.column.row
          [:div.column
           (doall (map (partial range-selector @trend-range) range-configs))]
          [:div.column
           [:button.ui.button.right.floated {:on-click #(rf/dispatch [::controller/removetrend active-trend-index])}
            [:i.trash.gray.icon]
            "Remove plot"]
           [:button.ui.button.right.floated {:on-click #(rf/dispatch [::controller/untrend active-trend-index])}
            [:i.eye.slash.gray.icon]
            "Remove variables from plot"]]]
         [:div.one.column.row
          [trend-inner {:trend-values (format-data-for-plotting plot-type trend-values)
                        :plot-type    plot-type
                        :trend-id     id}]]]))))

(rf/reg-sub ::active-trend #(get-in % [:state :trends (-> % :active-trend-index int)]))

(rf/reg-sub ::trend-range
            :<- [::active-trend]
            #(-> % :spec :range))

(defn- ascending-points? [tuples]
  (= tuples
     (sort-by :x tuples)))

(s/def ::module string?)
(s/def ::signal string?)
(s/def ::trend-point (s/keys :req-un [::x ::y]))
(s/def ::ascending-points ascending-points?)
(s/def ::trend-data (s/and (s/coll-of ::trend-point :kind vector?) ::ascending-points))
(s/def ::trend-value (s/keys :req-un [::module ::signal ::trend-data]))
(s/def ::trend-values (s/coll-of ::trend-value))
