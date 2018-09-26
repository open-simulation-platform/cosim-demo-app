(ns cse-client.view
  (:require [cse-client.trend :as trend]
            [kee-frame.core :as k]
            [re-frame.core :as rf]
            [cse-client.controller :as controller]))

(defn module-listing []
  (let [{:keys [signals] :as module} @(rf/subscribe [:module])]
    [:div
     [:a {:href (k/path-for [:index])} "Back to modules"]
     [:table.ui.single.line.striped.selectable.table
      [:thead
       [:tr
        [:th "Name"]
        [:th "Type"]
        [:th "Value"]
        [:th "..."]]]
      [:tbody
       (map (fn [{:keys [name value]}]
              [:tr {:key (str "signal-" name)}
               [:td name]
               [:td "scalar"]
               [:td value]
               [:td [:a {:href (k/path-for [:trend {:module (:name module) :signal name}])} "Trend"]]])
            signals)]]]))

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

(defn root-comp []
  [:div
   [:div.ui.inverted.huge.borderless.fixed.fluid.menu
    [:a.header.item "Project name"]
    [:div.right.menu
     [:div.item
      [:div.ui.small.input
       [:input {:placeholder "Search..."}]]]
     [:a.item "Dashboard"]
     [:a.item "Settings"]]]
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
         :index [:div "Index here......"]
         nil [:div "Loading..."]]]]]]]
   #_[controls]
   ])
