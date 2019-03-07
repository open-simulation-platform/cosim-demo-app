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

(defn trend-layout []
  {:autosize           true
   :use-resize-handler true})

(defn scatter-layout []
  {:xaxis {:autorange true
           :autotick  true
           :ticks     ""}})

(defn layout-selector [plot-type]
  (case plot-type
    "trend" (trend-layout)
    "scatter" (scatter-layout)))

(defn create-traces [plot-type trend-values]
  (case plot-type
    "trend" trend-values
    "scatter" (map (fn [[xvals yvals]]
                     (merge xvals yvals)) (partition 2 trend-values))))

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

(defn delete-series [dom-node]
  (let [num-series (-> dom-node .-data .-length)]
    (doseq [_ (range num-series)]
      (js/Plotly.deleteTraces dom-node 0))))

(defn maybe-update-series [dom-node trend-values]
  (let [num-series (-> dom-node .-data .-length)]
    (when (not= num-series (count trend-values))
      (doseq [_ (range num-series)]
        (js/Plotly.deleteTraces dom-node 0))
      (doseq [trend-variable trend-values]
        (js/Plotly.addTraces dom-node (clj->js (new-series trend-variable)))))))

(defn update-chart-data [dom-node trend-values trend-id]
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

(defn- set-dom-element-height! [dom-node height]
  (-> dom-node .-style .-height (set! height)))

(defn trend-inner []
  (let [update (fn [comp]
                 (let [{:keys [trend-values trend-id]} (r/props comp)]
                   (update-chart-data (r/dom-node comp) trend-values trend-id)))]
    (r/create-class
      {:component-did-mount  (fn [comp]
                               (let [{:keys [trend-layout]} (r/props comp)
                                     dom-node (r/dom-node comp)
                                     _ (set-dom-element-height! dom-node plot-container-height)]
                                 (js/Plotly.react dom-node
                                                 (clj->js [{:x    []
                                                            :y    []
                                                            :mode "lines"
                                                            :type "scatter"}])
                                                 (clj->js trend-layout)
                                                 (clj->js {:responsive true}))
                                 (.on (r/dom-node comp) "plotly_relayout" relayout-callback)))
       :component-did-update update
       :reagent-render       (fn []
                               [:div.column])})))

(defn trend-outer []
  (let [trend-range (rf/subscribe [::trend-range])
        active-trend (rf/subscribe [::active-trend])]
    (fn []
      (let [{:keys [id plot-type label trend-values]} @active-trend]
        [:div.ui.one.column.grid
         [c/variable-override-editor nil nil label [::controller/set-label]]
         [:div.two.column.row
          (if-not (= "scatter" plot-type)
            [:div.column
             (doall (map (partial range-selector @trend-range) range-configs))]
            [:div.column])
          [:div.column
           [:button.ui.button.right.floated {:on-click #(rf/dispatch [::controller/removetrend])} "Remove trend"]
           [:button.ui.button.right.floated {:on-click #(rf/dispatch [::controller/untrend])} "Untrend all"]]]
         [:div.one.column.row
          [trend-inner {:trend-values (create-traces plot-type trend-values)
                        :trend-layout (layout-selector plot-type)
                        :trend-id     id}]]]))))

(rf/reg-sub ::trend-range :trend-range)
(rf/reg-sub ::active-trend #(get-in % [:state :trends (-> % :active-trend-index int)]))

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
