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

(def trend-layout
  {:xaxis              {:title "Time [s]"}
   :autosize           true
   :use-resize-handler true
   :showlegend         true
   :legend             {:orientation "h"}})

(def scatter-layout
  {:xaxis      {:autorange true
                :autotick  true
                :ticks     ""}
   :showlegend true
   :legend     {:orientation "h"}})

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

(defn- delete-series [dom-node]
  (let [num-series (-> dom-node .-data .-length)]
    (doseq [_ (range num-series)]
      (js/Plotly.deleteTraces dom-node 0))))

(defn- add-time-series-traces [dom-node plots]
  (let [plot-legend-title (fn [{:keys [module signal causality type]}]
                            (str/join " - " [module signal causality type]))]
    (doseq [plot plots]
      (js/Plotly.addTraces dom-node (clj->js {:name (plot-legend-title plot) :x [] :y []})))))

(defn- add-xy-plot-traces [dom-node plots]
  (let [plot-legend-title (fn [p]
                            (let [first-signal ((keyword (str first-signal-ns "/" 'signal)) p)
                                  second-signal ((keyword (str second-signal-ns "/" 'signal)) p)]
                              (str/join " / " [first-signal second-signal])))]
    (doseq [plot plots]
      (js/Plotly.addTraces dom-node (clj->js {:name (plot-legend-title plot) :x [] :y []})))))

(defn- maybe-update-series [dom-node trend-values]
  (let [num-series (-> dom-node .-data .-length)]
    (when (not= num-series (count trend-values))
      (doseq [_ (range num-series)]
        (js/Plotly.deleteTraces dom-node 0))
      (case (:plot-type @(rf/subscribe [::active-trend]))
        "trend" (add-time-series-traces dom-node trend-values)
        "scatter" (add-xy-plot-traces dom-node trend-values)))))

(defn- update-chart-data [dom-node trend-values trend-id trend-layout]
  (when-not (= trend-id @id-store)
    (reset! id-store trend-id)
    (delete-series dom-node))
  (s/assert ::trend-values trend-values)
  (let [init-data {:x [] :y []}
        data (reduce (fn [data {:keys [xvals yvals]}]
                       (-> data
                           (update :x conj xvals)
                           (update :y conj yvals)))
                     init-data trend-values)]
    (maybe-update-series dom-node trend-values)
    (js/Plotly.update dom-node (clj->js data) (clj->js trend-layout))))

(defn- relayout-callback [js-event]
  (let [event (js->clj js-event)
        begin (get event "xaxis.range[0]")
        end (get event "xaxis.range[1]")
        auto? (get event "xaxis.autorange")
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
  (let [update (fn [comp]
                 (let [{:keys [trend-values trend-id trend-layout]} (r/props comp)]
                   (update-chart-data (r/dom-node comp) trend-values trend-id trend-layout)))]
    (r/create-class
      {:component-did-mount  (fn [comp]
                               (let [{:keys [trend-layout]} (r/props comp)
                                     dom-node (r/dom-node comp)
                                     _ (set-dom-element-height! dom-node plot-container-height)]
                                 (js/Plotly.newPlot dom-node
                                                    (clj->js [{:x    []
                                                               :y    []
                                                               :mode "lines"
                                                               :type "scatter"}])
                                                    (clj->js trend-layout)
                                                    (clj->js {:responsive true}))
                                 (.on dom-node "plotly_relayout" relayout-callback)))
       :component-did-update update
       :reagent-render       (fn []
                               [:div.column])})))

(defn trend-outer []
  (let [trend-range (rf/subscribe [::trend-range])
        active-trend (rf/subscribe [::active-trend])
        active-trend-index (rf/subscribe [:active-trend-index])]
    (fn []
      (let [{:keys [id plot-type label trend-values]} @active-trend
            active-trend-index (int @active-trend-index)
            name (plot-type-from-label label)]
        [:div.ui.one.column.grid
         [c/variable-override-editor nil nil name [::controller/set-label]]
         [:div.two.column.row
          [:div.column
           (doall (map (partial range-selector @trend-range) range-configs))]
          [:div.column
           [:button.ui.button.right.floated {:on-click #(rf/dispatch [::controller/removetrend active-trend-index])}
            [:i.trash.gray.icon]
            "Remove trend"]
           [:button.ui.button.right.floated {:on-click #(rf/dispatch [::controller/untrend active-trend-index])}
            [:i.eye.slash.gray.icon]
            "Remove variables from trend"]]]
         [:div.one.column.row
          [trend-inner {:trend-values (format-data-for-plotting plot-type trend-values)
                        :trend-layout (layout-selector plot-type)
                        :trend-id     id}]]]))))

(rf/reg-sub ::trend-range :trend-range)
(rf/reg-sub ::active-trend #(get-in % [:state :trends (-> % :active-trend-index int)]))

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
