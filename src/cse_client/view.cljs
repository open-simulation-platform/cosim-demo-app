(ns cse-client.view
  (:require [cse-client.trend :as trend]
            [kee-frame.core :as k]
            [re-frame.core :as rf]
            [reagent.core :as r]
            [cse-client.controller :as controller]
            [cse-client.dp :as dp]
            [cse-client.config :refer [socket-url]]))

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
               [:td [:a {:href (k/path-for [:trend {:module (:name module) :signal name}])} "Trend"]]])
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

(defn index-page []
  (let [loaded? (rf/subscribe [:loaded?])
        load-dir (r/atom "")]
    (fn []
      (if @loaded?
        [dp/svg-component]
        [:div [:div "Specify a folder with FMUs:"]
         [:div.row
          [:input {:style         {:min-width "400px"}
                   :type          :text
                   :default-value ""
                   :on-change     #(reset! load-dir (-> % .-target .-value))}]]
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
       [:div.item
        [:div "Connection:" (:state @socket-state)]]
       [:button.ui.button {:on-click #(rf/dispatch [::controller/play])
                           :disabled (or (not @loaded?) (= @status "play"))} "Play"]
       [:button.ui.button {:on-click #(rf/dispatch [::controller/pause])
                           :disabled (or (not @loaded?) (= @status "pause"))} "Pause"]]]
     [:div.ui.grid
      [:div.row
       [:div#sidebar.column
        [sidebar]]
       [:div#content.column
        [:div.ui.grid
         [:div.row
          [:h1.ui.huge.header "Dashboard"]]
         [:div.ui.divider]
         [:div.row
          [k/switch-route (comp :name :data)
           :trend [trend/trend-outer]
           :module [module-listing]
           :index [index-page]
           nil [:div "Loading..."]]]]]]]]))
