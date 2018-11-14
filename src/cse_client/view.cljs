(ns cse-client.view
  (:require [cse-client.trend :as trend]
            [kee-frame.core :as k]
            [re-frame.core :as rf]
            [reagent.core :as r]
            [cse-client.controller :as controller]
            [cse-client.config :refer [socket-url]]))

(goog-define default-load-dir "")

(defn tab-content [tabby]
  (let [module @(rf/subscribe [:module])
        signals @(rf/subscribe [:signals])
        active @(rf/subscribe [:active-causality])]
    [:div.ui.bottom.attached.tab.segment {:data-tab tabby
                                          :class    (when (= tabby active) "active")}
     [:table.ui.single.line.striped.selectable.table
      [:thead
       [:tr
        [:th "Name"]
        [:th "Type"]
        [:th "Value"]
        [:th "..."]]]
      [:tbody
       (map (fn [{:keys [name value causality type]}]
              [:tr {:key (str causality "-" name)}
               [:td name]
               [:td type]
               [:td value]
               [:td [:a {:href (k/path-for [:trend {:module (:name module) :signal name :causality causality :type type}])} "Trend"]]])
            signals)]]]))

(defn module-listing []
  (let [causalities @(rf/subscribe [:causalities])
        active @(rf/subscribe [:active-causality])]
    [:div
     [:div.ui.top.attached.tabular.menu
      (for [causality causalities]
        ^{:key (str "tab-" causality)}
        [:a.item {:data-tab causality
                  :class    (when (= causality active) "active")
                  :on-click #(rf/dispatch [::controller/causality-enter causality])}
         causality])]
     (for [causality causalities]
       ^{:key (str "tab-content-" causality)}
       [tab-content causality])]))

(defn sidebar []
  (let [modules @(rf/subscribe [:modules])
        route @(rf/subscribe [:kee-frame/route])
        route-name (-> route :data :name)
        route-module (-> route :path-params :module)]
    [:div.ui.secondary.vertical.fluid.menu
     [:a.item {:href  (k/path-for [:index])
               :class (when (= route-name :index) :active)} "Overview"]
     (map (fn [module]
            [:a.item {:class (when (= route-module module) :active)
                      :key   module
                      :href  (k/path-for [:module {:module module}])} module])
          modules)]))

(defn controls []
  [:div.ui.buttons
   [:button.ui.button {:on-click #(rf/dispatch [::controller/play])} "Play"]
   [:button.ui.button {:on-click #(rf/dispatch [::controller/pause])} "Pause"]])

(defn dashboard []
  [:table.ui.basic.table.definition
   [:tbody
    (for [[k v] @(rf/subscribe [:overview])]
      ^{:key k}
      [:tr
       [:td k]
       [:td v]])]])

(defn index-page []
  (let [loaded? (rf/subscribe [:loaded?])
        load-dir (r/atom default-load-dir)]
    (fn []
      (if @loaded?
        [dashboard]
        [:div [:div "Specify a folder with FMUs:"]
         [:div.row
          [:input {:style     {:min-width "400px"}
                   :type      :text
                   :value     @load-dir
                   :on-change #(reset! load-dir (-> % .-target .-value))}]]
         [:div.row [:button.ui.button.pull-right {:disabled (empty? @load-dir)
                                                  :on-click #(rf/dispatch [::controller/load @load-dir])} "Load simulation"]]]))))

(defn root-comp []
  (let [socket-state (rf/subscribe [:kee-frame.websocket/state socket-url])
        loaded? (rf/subscribe [:loaded?])
        status (rf/subscribe [:status])]
    [:div
     [:div.ui.inverted.huge.borderless.fixed.fluid.menu
      [:a.header.item "Open Simulation Platform"]
      [:div.right.menu
       (when (and @loaded? (= @status "pause"))
         [:a.item {:on-click #(rf/dispatch [::controller/teardown])} "Teardown"])
       (when (= :disconnected (:state @socket-state))
         [:div.item
          [:div "Lost server connection!"]])
       [:div.item
        [:div "Time: " @(rf/subscribe [:time])]]
       (when (and @loaded? (= @status "pause"))
         [:a.item {:on-click #(rf/dispatch [::controller/play])} "Play"])
       (when (and @loaded? (= @status "play"))
         [:a.item {:on-click #(rf/dispatch [::controller/pause])} "Pause"])]]
     [:div.ui.grid
      [:div.row
       [:div#sidebar.column
        [sidebar]]
       [:div#content.column
        [:div.ui.grid
         [:div.row
          [:h1.ui.huge.header [k/switch-route (comp :name :data)
                               :trend "Trend"
                               :module "Model"
                               :index "Simulation status"
                               nil [:div "Loading..."]]]]
         [:div.ui.divider]
         [:div.row
          [k/switch-route (comp :name :data)
           :trend [trend/trend-outer]
           :module [module-listing]
           :index [index-page]
           nil [:div "Loading..."]]]]]]]]))
