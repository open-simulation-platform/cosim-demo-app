(ns cse-client.view
  (:require [cse-client.trend :as trend]
            [kee-frame.core :as k]
            [re-frame.core :as rf]
            [cse-client.controller :as controller]))

(defn module-listing [signals module-name]
  [:div
   [:a {:href (k/path-for [:index])} "Back to modules"]
   [:ul
    (map (fn [signal]
           [:li {:key (str name "_ " (:name signal))} (:name signal) ": " (:value signal)
            [:a {:href (k/path-for [:trend {:module module-name :signal (:name signal)}])} "Trend"]])
         signals)]])

(defn modules-menu [modules]
  [:div
   [:h3 "Modules"]
   [:ul
    (map (fn [module]
           [:li {:key module} [:a {:href (k/path-for [:module {:name module}])} module]])
         modules)]])

(defn controls []
  [:div.ui.buttons
   [:button.ui.button {:on-click #(rf/dispatch [::controller/play])} "Play"]
   [:button.ui.button {:on-click #(rf/dispatch [::controller/pause])} "Pause"]])

(defn root-comp []
  (let [{:keys [module modules]} @(rf/subscribe [:state])
        {:keys [name signals]} module]
    [:div
     [controls]
     [k/switch-route (comp :name :data)
      :trend [trend/trend-outer]
      :module [module-listing signals name]
      :index [modules-menu modules]
      nil [:div "Loading..."]]]))
