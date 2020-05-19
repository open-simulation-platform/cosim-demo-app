;; This Source Code Form is subject to the terms of the Mozilla Public
;; License, v. 2.0. If a copy of the MPL was not distributed with this
;; file, You can obtain one at https://mozilla.org/MPL/2.0/.

(ns cse-client.trend
  (:require [cse-client.controller :as controller]
            [reagent.core :as r]
            [cljsjs.plotly]
            [re-frame.core :as rf]
            [cljs.spec.alpha :as s]
            [cse-client.components :as c]
            [clojure.string :as str]))

(def id-store (atom nil))

(def plot-heights {:collapsed "50vh"
                   :expanded  "75vh"})

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
   :margin             {:l 0 :r 0 :t 25 :pad 5}
   :xaxis              {:automargin true}
   :yaxis              {:automargin true}
   :legend             {:orientation "h"
                        :uirevision true}})

(def trend-layout
  (merge common-layout
         {:xaxis {:automargin true
                  :title {:text "Time [s]"
                          :font {:size 14}}}}))

(def scatter-layout
  (merge common-layout
         {:xaxis {:autorange true
                  :autotick  true
                  :ticks     ""
                  :automargin true}}))

(def options
  {:responsive true
   :toImageButtonOptions {:width 1280 :height 768}})

(defn- plotly-expand-button [plot-height plot-expanded?]
  (if @plot-expanded?
    [:button.ui.button.right.floated {:on-click (fn []
                                                  (rf/dispatch [::controller/set-plot-height (:collapsed plot-heights)])
                                                  (swap! plot-expanded? not))}
     [:i.compress.icon]
     "Compress plot"]
    [:button.ui.button.right.floated {:on-click (fn []
                                                  (rf/dispatch [::controller/set-plot-height (:expanded plot-heights)])
                                                  (swap! plot-expanded? not))}
     [:i.expand.icon]
     "Expand plot"]))

(defn- layout-selector [plot-type]
  (case plot-type
    "trend"   trend-layout
    "scatter" scatter-layout
    {}))

(defn- autoscale! []
  (js/Plotly.relayout
   "plotly"
   (clj->js {:xaxis {:autorange true}
             :yaxis {:autorange true}})))

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
    "trend"   trend-values
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

(defn- update-traces [dom-node trend-values]
  (let [num-series (-> dom-node .-data .-length)]
    (doseq [_ (range num-series)]
      (js/Plotly.deleteTraces dom-node 0))
    (case (:plot-type @(rf/subscribe [::active-trend]))
      "trend" (add-traces dom-node trend-values time-series-legend-name)
      "scatter" (add-traces dom-node trend-values xy-plot-legend-name))))

(defn- update-chart-data [dom-node trend-values layout trend-id]
  (when-not (= trend-id @id-store)
    (reset! id-store trend-id)
    (delete-series dom-node)
    (autoscale!))
  (s/assert ::trend-values trend-values)
  (let [init-data {:x [] :y []}
        data      (reduce (fn [data {:keys [xvals yvals]}]
                            (-> data
                                (update :x conj xvals)
                                (update :y conj yvals)))
                          init-data trend-values)]
    (update-traces dom-node trend-values)
    (js/Plotly.update dom-node (clj->js data) (clj->js layout) (clj->js options))))

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
                      (let [{:keys [trend-values trend-id plot-type plot-height]} (r/props comp)
                            dom-node            (r/dom-node comp)
                            _                   (set-dom-element-height! dom-node plot-height)
                            layout                                    (layout-selector plot-type)]
                        (update-chart-data (r/dom-node comp) trend-values layout trend-id)))
        render-plot (fn [comp]
                      (let [{:keys [plot-type plot-height]} (r/props comp)
                            dom-node            (r/dom-node comp)
                            _                   (set-dom-element-height! dom-node plot-height)
                            layout              (layout-selector plot-type)]
                        (js/Plotly.newPlot dom-node
                                           (clj->js [{:x    []
                                                      :y    []
                                                      :mode "lines"
                                                      :type "scatter"}])
                                           (clj->js layout)
                                           (clj->js options))))]
    (r/create-class
     {:component-did-mount  (fn [comp]
                              (render-plot comp)
                              (.on (r/dom-node comp) "plotly_relayout" relayout-callback))
      :component-did-update update-plot
      :reagent-render       (fn []
                              [:div#plotly.column])})))

(defn variable-row []
  (let [untrending? (r/atom false)]
    (fn [trend-idx module signal causality val]
      [:tr
       [:td module]
       [:td signal]
       [:td causality]
       [:td (when (and (some? val) (number? val))
              (.toFixed val 4))]
       #_[:td
          (if @untrending?
            [:i.fa.fa-spinner.fa-spin]
            [:span {:style         {:float 'right :cursor 'pointer}
                    :data-tooltip  "Remove variable from plot"
                    :data-position "top center"}
             [:i.eye.slash.gray.icon {:on-click #(rf/dispatch [::controller/untrend-single trend-idx (str module "." signal)])}]])]])))

(defn variables-table [trend-idx trend-values]
  [:table.ui.single.line.striped.table
   [:thead
    [:tr
     [:th "Model"]
     [:th "Variable"]
     [:th "Causality"]
     [:th "Value"]
     #_[:th {:style {:text-align 'right}} "Remove"]]]
   [:tbody
    (doall
     (for [{:keys [module signal causality yvals]} trend-values] ^{:key (str module signal (rand-int 9999))}
       [variable-row trend-idx module signal causality (last yvals)]))]])

(defn trend-outer []
  (let [trend-range        (rf/subscribe [::trend-range])
        active-trend       (rf/subscribe [::active-trend])
        active-trend-index (rf/subscribe [:active-trend-index])
        plot-height        (rf/subscribe [:plot-height])
        plot-expanded?     (r/atom false)]
    (fn []
      (let [{:keys [id plot-type label trend-values]} @active-trend
            active-trend-index                        (int @active-trend-index)
            name                                      (plot-type-from-label label)]
        [:div.ui.one.column.grid

         [c/text-editor name [::controller/set-label] "Rename plot"]

         [:div.two.column.row
          [:div.column
           (doall (map (partial range-selector @trend-range) range-configs))]
          [:div.column
           [plotly-expand-button plot-height plot-expanded?]
           [:button.ui.button.right.floated {:on-click #(rf/dispatch [::controller/removetrend active-trend-index])}
            [:i.trash.gray.icon]
            "Remove plot"]
           [:button.ui.button.right.floated {:on-click #(rf/dispatch [::controller/untrend active-trend-index])}
            [:i.eye.slash.gray.icon]
            "Remove all variables from plot"]]]

         [:div.one.column.row
          [trend-inner {:trend-values (format-data-for-plotting plot-type trend-values)
                        :plot-type    plot-type
                        :plot-height  (or @plot-height (:collapsed plot-heights))
                        :trend-id     id}]]

         (when (not @plot-expanded?)

           [variables-table active-trend-index trend-values])]))))

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
