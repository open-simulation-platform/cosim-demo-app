(ns cse-client.scenario
  (:require [re-frame.core :as rf]
            [kee-frame.core :as k]
            [clojure.string :as str]
            [cse-client.controller :as controller]))

(defn scenario-filename-to-name [file-name]
  (str/capitalize (first (str/split file-name "."))))

(defn cellie [text valid? message]
  (if valid?
    [:td text]
    [:td {:class        "error"
          :data-tooltip message} text]))

(defn run-button [scenario-id any-running? invalid?]
  [:button.ui.right.labeled.icon.button
   {:disabled     (or any-running? invalid?)
    :on-click     #(rf/dispatch [::controller/load-scenario scenario-id])}
   (if invalid? [:i.red.ban.icon] [:i.green.play.icon])
   (if invalid?
     "Invalid syntax"
     (if any-running? "Other scenario running" "Load scenario"))])

(defn running-button [scenario-id]
  [:button.ui.right.labeled.icon.button
   {:on-click #(rf/dispatch [::controller/abort-scenario scenario-id])}
   "Abort scenario"
   [:i.red.stop.icon]])

(defn one []
  (let [scenario @(rf/subscribe [:scenario])
        scenario-id @(rf/subscribe [:scenario-id])
        running? @(rf/subscribe [:scenario-running? scenario-id])
        any-running? @(rf/subscribe [:any-scenario-running?])]
    [:div
     [:div.ui.header "Actions"]
     (if running?
       [running-button scenario-id]
       [run-button scenario-id any-running? (not (:valid? scenario))])
     [:div.ui.header "Description"]
     [:div (or (:description scenario) "No description available")]
     [:div.ui.header "Events"]
     [:table.ui.celled.striped.selectable.fluid.table
      [:thead
       [:tr
        [:th "Time"]
        [:th "Model"]
        [:th "Variable"]
        [:th "Type"]
        [:th "Causality"]
        [:th "Action"]
        [:th "Value"]]]
      [:tbody
       (map-indexed (fn [index {:keys [time model variable causality type action value model-valid? variable-valid? validation-message] :as event}]
                      [:tr {:key (str "scenario-" index "-event")}
                       [:td time]
                       [cellie model model-valid? validation-message]
                       [cellie variable variable-valid? validation-message]
                       [cellie type variable-valid? validation-message]
                       [cellie causality variable-valid? validation-message]
                       [:td action]
                       [:td value]])
                    (:events scenario))]]
     [:div.ui.header "End time"]
     [:div (or (:end scenario)
               (-> scenario :events last :time))]]))

(defn overview []
  (let [scenarios @(rf/subscribe [:scenarios])]
    [:div.ui.large.list
     (map (fn [{:keys [id running?]}]
            [:div.item {:key (str "scenario-" id)}
             [:i.file.alternate.icon {:class (when running? "green")}]
             [:div.content
              [:a {:href (k/path-for [:scenario {:id id}])} (str (scenario-filename-to-name id) " - " id)]]])
          scenarios)]))
