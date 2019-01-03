(ns cse-client.view
  (:require [cse-client.trend :as trend]
            [kee-frame.core :as k]
            [re-frame.core :as rf]
            [reagent.core :as r]
            [cse-client.controller :as controller]
            [cse-client.config :refer [socket-url]]))

(goog-define default-load-dir "")
(goog-define default-log-dir "")

(defn variable-override-editor [module {:keys [name causality type value]}]
  (let [editing? (r/atom false)
        internal-value (r/atom value)]
    (fn [_ {:keys [value]}]
      (if @editing?
        [:div.ui.action.input.fluid
         [:input {:type      :text
                  :autoFocus true
                  :id        (str "input-" name)
                  :value     (if @editing? @internal-value value)
                  :on-change #(reset! internal-value (.. % -target -value))}]
         [:button.ui.right.icon.button
          {:on-click (fn [_]
                       (rf/dispatch [::controller/set-value module name causality type @internal-value])
                       (reset! editing? false))}
          [:i.check.link.icon]]
         [:button.ui.right.icon.button
          {:on-click #(reset! editing? false)}
          [:i.times.link.icon]]]
        [:div {:style    {:cursor :pointer}
               :on-click (fn [_]
                           (reset! editing? true)
                           (reset! internal-value value))}
         value]))))

(defn variable-display [module {:keys [value editable?] :as variable}]
  (if editable?
    [variable-override-editor module variable]
    [:div value]))

(defn tab-content [tabby]
  (let [module @(rf/subscribe [:module])
        signals @(rf/subscribe [:signals])
        active @(rf/subscribe [:active-causality])]
    [:div.ui.bottom.attached.tab.segment {:data-tab tabby
                                          :class    (when (= tabby active) "active")}
     [:table.ui.single.line.striped.selectable.fixed.table
      [:thead
       [:tr
        [:th "Name"]
        [:th "Type"]
        [:th "Value"]
        [:th "..."]]]
      [:tbody
       (map (fn [{:keys [name value causality type] :as variable}]
              [:tr {:key (str (:name module) "-" causality "-" name)}
               [:td name]
               [:td type]
               [:td [variable-display (:name module) variable]]
               [:td [:a {:style    {:cursor :pointer}
                         :on-click #(rf/dispatch [::controller/add-to-trend (:name module) name causality type])} "Add to trend"]]])
            signals)]]]))

(defn module-listing []
  (let [causalities @(rf/subscribe [:causalities])
        active @(rf/subscribe [:active-causality])]
    [:div.ui.one.column.grid
     [:div.one.column.row
      [:div.column
       [:div.ui.top.attached.tabular.menu
        (for [causality causalities]
          ^{:key (str "tab-" causality)}
          [:a.item {:data-tab causality
                    :class    (when (= causality active) "active")
                    :on-click #(rf/dispatch [::controller/causality-enter causality])}
           causality])]
       (for [causality causalities]
         ^{:key (str "tab-content-" causality)}
         [tab-content causality])]]]))

(defn sidebar []
  (let [modules @(rf/subscribe [:modules])
        route @(rf/subscribe [:kee-frame/route])
        route-name (-> route :data :name)
        route-module (-> route :path-params :module)
        loaded? @(rf/subscribe [:loaded?])
        trend-count @(rf/subscribe [:trend-count])]
    [:div.ui.secondary.vertical.fluid.menu
     [:a.item {:href  (k/path-for [:index])
               :class (when (= route-name :index) :active)}
      "Overview"]
     (when loaded?
       [:a.item {:href  (k/path-for [:trend])
                 :class (when (= route-name :trend) :active)}
        "Trend"
        [:div.ui.teal.left.pointing.label trend-count]])
     [:div.ui.divider]
     (map (fn [module]
            [:a.item {:class (when (= route-module module) :active)
                      :key   module
                      :href  (k/path-for [:module {:module module}])} module])
          modules)]))

(defn realtime-button []
  (if @(rf/subscribe [:realtime?])
    [:button.ui.button
     {:on-click     #(rf/dispatch [::controller/disable-realtime])
      :data-tooltip "Execute simulation as fast as possible"}
     "Disable"]
    [:button.ui.button
     {:on-click     #(rf/dispatch [::controller/enable-realtime])
      :data-tooltip "Execute simulation towards real time target"}
     "Enable"]))

(defn teardown-button []
  [:button.ui.button {:on-click     #(rf/dispatch [::controller/teardown])
                      :disabled     (not (= @(rf/subscribe [:status]) "pause"))
                      :data-tooltip "Tear down current simulation, allowing for a simulation restart. Simulation must be paused."}
   "Tear down"])

(defn dashboard []
  [:div
   [:table.ui.basic.table.definition
    [:tbody
     (for [[k v] @(rf/subscribe [:overview])]
       ^{:key k}
       [:tr
        [:td k]
        [:td v]])]]
   [:div.header "Controls"]
   [:table.ui.basic.table.definition
    [:tbody
     [:tr [:td "Real time target"] [:td [realtime-button]]]
     [:tr [:td "Simulation execution"] [:td [teardown-button]]]]]])

(defn index-page []
  (let [loaded? (rf/subscribe [:loaded?])
        load-dir (r/atom default-load-dir)
        log-dir (r/atom default-log-dir)]
    (fn []
      (if @loaded?
        [dashboard]
        [:div.ui.two.column.grid
         [:div.two.column.row
          [:div.column
           [:div.ui.fluid.right.labeled.input {:data-tooltip "Specify a directory containing FMUs (and optionally a SystemStructure.ssd file)"}
            [:input {:style       {:min-width "400px"}
                     :type        :text
                     :placeholder "Load folder..."
                     :value       @load-dir
                     :on-change   #(reset! load-dir (-> % .-target .-value))}]
            [:div.ui.label "FMUs"]]]]
         [:div.two.column.row
          [:div.column
           [:div.ui.fluid.right.labeled.input {:data-tooltip "[Optional] Specify a directory where output log files will be stored"}
            [:input {:style       {:min-width "400px"}
                     :type        :text
                     :placeholder "Log folder... (optional)"
                     :value       @log-dir
                     :on-change   #(reset! log-dir (-> % .-target .-value))}]
            [:div.ui.label "logs"]]]]
         [:div.two.column.row
          [:div.column
           [:button.ui.button.pull-right {:disabled (empty? @load-dir)
                                          :on-click #(rf/dispatch [::controller/load @load-dir @log-dir])} "Load simulation"]]]]))))

(defn root-comp []
  (let [socket-state (rf/subscribe [:kee-frame.websocket/state socket-url])
        loaded? (rf/subscribe [:loaded?])
        status (rf/subscribe [:status])]
    [:div
     [:div.ui.inverted.huge.borderless.fixed.menu
      [:a.header.item {:href "/"} "Core Simulation Environment - demo application"]
      [:div.right.menu
       (when (= :disconnected (:state @socket-state))
         [:div.item
          [:div "Lost server connection!"]])
       (when @loaded?
         [:div.item
          [:div "Time: " @(rf/subscribe [:time])]])
       (when (and @loaded? (= @status "pause"))
         [:a.item {:on-click #(rf/dispatch [::controller/play])} "Play"])
       (when (and @loaded? (= @status "play"))
         [:a.item {:on-click #(rf/dispatch [::controller/pause])} "Pause"])
       [:div.ui.simple.dropdown.item
        [:i.ellipsis.vertical.icon]
        [:div.menu
         [:a.item [:i.question.circle.icon] "User guide"]
         [:a.item {:href "mailto:issue@opensimulationplatform.com?subject=Feedback to CSE Team"} [:i.mail.icon] "Provide feedback"]]]]]
     [:div.ui.grid
      [:div.row
       [:div#sidebar.column
        [sidebar]]
       [:div#content.column
        [:div.ui.grid
         [:div.row
          [:h1.ui.huge.header [k/switch-route (comp :name :data)
                               :module "Model"
                               :trend "Trend"
                               :index (if @loaded? "Simulation status" "Simulation setup")
                               nil [:div "Loading..."]]]]
         [:div.ui.divider]
         [:div.row
          [k/switch-route (comp :name :data)
           :trend [trend/trend-outer]
           :module [module-listing]
           :index [index-page]
           nil [:div "Loading..."]]]]]]]
     (when (= :disconnected (:state @socket-state))
       [:div.ui.page.dimmer.transition.active
        {:style {:display :flex}}
        [:div.content
         [:div.center
          [:h2.ui.inverted.icon.header
           [:i.heartbeat.icon]
           "Lost server connection!"]
          [:div.sub.header "It looks like the server is down. Try restarting the server and hit F5"]]]])]))
