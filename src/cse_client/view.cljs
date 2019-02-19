(ns cse-client.view
  (:require [cse-client.trend :as trend]
            [kee-frame.core :as k]
            [re-frame.core :as rf]
            [reagent.core :as r]
            [cse-client.controller :as controller]
            [cse-client.config :refer [socket-url]]
            [cse-client.guide :as guide]
            [cse-client.components :as c]
            [clojure.string :as str]
            [fulcrologic.semantic-ui.factories :as semantic]
            [fulcrologic.semantic-ui.icons :as icons]))

(goog-define default-load-dir "")
(goog-define default-log-dir "")

(defn variable-display [module {:keys [name causality type editable?] :as variable}]
  (let [value @(rf/subscribe [:signal-value module name causality type])]
    (if editable?
      [c/variable-override-editor module variable value]
      [:div value])))

(defn trend-item [current-module name causality type value-reference {:keys [id index count label plot-type]}]
  (case plot-type
    "trend" (semantic/ui-dropdown-item
              {:key     (str "trend-item-" id)
               :text    label
               :label   (str/capitalize plot-type)
               :onClick #(rf/dispatch [::controller/add-to-trend current-module name causality type value-reference index])})
    "scatter"
    #_(semantic/ui-dropdown-item
        nil
        (semantic/ui-dropdown
          {:text  label
           :label (str/capitalize plot-type)}
          (semantic/ui-dropdown-menu
            nil
            (semantic/ui-dropdown-header nil "Add signal")
            (semantic/ui-dropdown-item {:text "Add to x"})
            (semantic/ui-dropdown-item {:text "Add to y"}))))
    (semantic/ui-dropdown-item
      {:key     (str "trend-item-" id)
       :text    label
       :label   (str/capitalize plot-type)
       :onClick #(rf/dispatch [::controller/add-to-trend current-module name causality type value-reference index])})))

(defn action-dropdown [current-module name causality type value-reference trend-info]
  (let [default-label (str "Trend #" (-> trend-info count inc))]
    (semantic/ui-dropdown
      {:icon   "chart line icon"
       :button true}
      (semantic/ui-dropdown-menu
        nil
        (semantic/ui-dropdown-header nil "Create new trend")
        (semantic/ui-dropdown-divider)
        (semantic/ui-dropdown-item
          {:text    "New time series"
           :onClick #(rf/dispatch [::controller/new-trend "trend" default-label])})
        (semantic/ui-dropdown-item
          {:text    "New scatter plot"
           :onClick #(rf/dispatch [::controller/new-trend "scatter" default-label])})
        (when-not (empty? trend-info)
          [(semantic/ui-dropdown-divider {:key "divider-1"})
           (semantic/ui-dropdown-header {:key (gensym "dropdown-trend-header")} "Add to trend")
           (semantic/ui-dropdown-divider {:key "divider-2"})
           (map (partial trend-item current-module name causality type value-reference) trend-info)])))))

(defn pages-menu []
  (let [current-page @(rf/subscribe [:current-page])
        pages @(rf/subscribe [:pages])
        vars-per-page @(rf/subscribe [:vars-per-page])]
    [:div.right.menu
     [:div.item
      [:div.ui.transparent.input
       [:button.ui.icon.button
        {:disabled (= 1 current-page)
         :on-click #(when (< 1 current-page)
                      (rf/dispatch [::controller/set-page (dec current-page)]))}
        [:i.chevron.left.icon]]
       [:input {:type     :text
                :readOnly true
                :value    (str "Page " current-page " of " (last pages))}]
       [:button.ui.icon.button
        {:disabled (= (last pages) current-page)
         :on-click #(when (< current-page (last pages))
                      (rf/dispatch [::controller/set-page (inc current-page)]))}
        [:i.chevron.right.icon]]]
      [:div.ui.icon.button.simple.dropdown
       [:i.sliders.horizontal.icon]
       [:div.menu
        [:div.header (str vars-per-page " variables per page")]
        [:div.item
         [:div.ui.icon.buttons
          [:button.ui.button {:on-click #(rf/dispatch [::controller/set-vars-per-page (- vars-per-page 5)])} [:i.minus.icon] 5]
          [:button.ui.button {:on-click #(rf/dispatch [::controller/set-vars-per-page (dec vars-per-page)])} [:i.minus.icon] 1]
          [:button.ui.button {:on-click #(rf/dispatch [::controller/set-vars-per-page (inc vars-per-page)])} [:i.plus.icon] 1]
          [:button.ui.button {:on-click #(rf/dispatch [::controller/set-vars-per-page (+ vars-per-page 5)])} [:i.plus.icon] 5]]]]]]]))

(defn tab-content [tabby]
  (let [current-module @(rf/subscribe [:current-module])
        module-signals @(rf/subscribe [:module-signals])
        active @(rf/subscribe [:active-causality])
        trend-info @(rf/subscribe [:trend-info])]
    [:div.ui.bottom.attached.tab.segment {:data-tab tabby
                                          :class    (when (= tabby active) "active")}
     [:table.ui.compact.single.line.striped.selectable.table
      [:thead
       [:tr
        [:th "Name"]
        [:th.one.wide "Type"]
        [:th "Value"]
        [:th.two.wide "Actions"]]]
      [:tbody
       (map (fn [{:keys [name causality type value-reference] :as variable}]
              [:tr {:key (str current-module "-" causality "-" name)}
               [:td name]
               [:td type]
               [:td [variable-display current-module variable]]
               [:td (action-dropdown current-module name causality type value-reference trend-info)]])
            module-signals)]]]))

(defn module-listing []
  (let [causalities @(rf/subscribe [:causalities])
        active @(rf/subscribe [:active-causality])
        module-active? @(rf/subscribe [:module-active?])
        current-module @(rf/subscribe [:current-module])]
    (if module-active?
      [:div.ui.one.column.grid
       [:div.one.column.row
        [:div.column
         [:div.ui.top.attached.tabular.menu
          (for [causality causalities]
            ^{:key (str "tab-" causality)}
            [:a.item {:data-tab causality
                      :class    (when (= causality active) "active")
                      :href     (k/path-for [:module {:module current-module :causality causality}])}
             causality])
          [pages-menu]]
         (for [causality causalities]
           ^{:key (str "tab-content-" causality)}
           [tab-content causality])]]]
      [:div.ui.active.centered.inline.text.massive.loader
       {:style {:margin-top "20%"}}
       "Loading"])))

(defn sidebar []
  (let [module-routes @(rf/subscribe [:module-routes])
        route @(rf/subscribe [:kee-frame/route])
        route-name (-> route :data :name)
        route-module (-> route :path-params :module)
        loaded? @(rf/subscribe [:loaded?])
        trend-info @(rf/subscribe [:trend-info])
        active-trend-index @(rf/subscribe [:active-trend-index])]
    [:div.ui.secondary.vertical.fluid.menu
     [:a.item {:href  (k/path-for [:index])
               :class (when (= route-name :index) :active)}
      "Overview"]
     (when (and loaded? (> (count trend-info) 0))
       [:div.item
        [:div.header "Trends"]
        [:div.menu
         (map (fn [{:keys [index label]}]
                [:a.item {:class (when (= index (int active-trend-index)) :active)
                          :key   label
                          :href  (k/path-for [:trend {:index index}])} label
                 [:div.ui.teal.left.pointing.label (-> trend-info (nth index) :count)]]) trend-info)]])
     [:div.ui.divider]
     [:div.item
      [:div.header "Models"]
      [:div.menu
       (map (fn [{:keys [name causality]}]
              [:a.item {:class (when (= route-module name) :active)
                        :key   name
                        :href  (k/path-for [:module {:module name :causality causality}])} name])
            module-routes)]]]))

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

(defn command-feedback-message []
  (when-let [{:keys [command message success]} @(rf/subscribe [:feedback-message])]
    [:div.ui.message
     {:class (if success "positive" "negative")
      :style {:position :absolute
              :bottom   20
              :right    20}}
     [:i.close.icon {:on-click #(rf/dispatch [::controller/close-feedback-message])}]
     [:div.header (if success
                    "Command success"
                    "Command failure")]
     [:p (str "Command: " command)]
     (when-not (str/blank? message)
       [:p (str "Message: " message)])]))

(defn dashboard []
  [:div
   [:table.ui.basic.table.definition
    [:tbody
     (for [[k v] @(rf/subscribe [:overview])]
       ^{:key k}
       [:tr
        [:td k]
        [:td v]])]]
   [:h3 "Controls"]
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
           [:button.ui.button.right.floated {:disabled (empty? @load-dir)
                                             :on-click #(rf/dispatch [::controller/load @load-dir @log-dir])} "Load simulation"]]]]))))

(defn root-comp []
  (let [socket-state (rf/subscribe [:kee-frame.websocket/state socket-url])
        loaded? (rf/subscribe [:loaded?])
        status (rf/subscribe [:status])
        module (rf/subscribe [:current-module])]
    [:div
     [:div.ui.inverted.huge.borderless.fixed.menu
      [:a.header.item {:href "/"} "Core Simulation Environment - demo application"]
      [:div.right.menu
       (when (= :disconnected (:state @socket-state))
         [:div.item
          [:div "Lost server connection!"]])
       (when @loaded?
         [:div.item
          [:div "RTF: " @(rf/subscribe [:real-time-factor])]])
       (when @loaded?
         [:div.item
          [:div "Time: " @(rf/subscribe [:time])]])
       (when (and @loaded? (= @status "pause"))
         [:a.item {:on-click #(rf/dispatch [::controller/play])} "Play"])
       (when (and @loaded? (= @status "play"))
         [:a.item {:on-click #(rf/dispatch [::controller/pause])} "Pause"])
       [:div.ui.simple.dropdown.item
        [:i.question.circle.icon]
        [:div.menu
         [:a.item {:href "/guide"} [:i.file.alternate.icon] "User guide"]
         [:a.item {:href "mailto:issue@opensimulationplatform.com?subject=Feedback to CSE Team"} [:i.mail.icon] "Provide feedback"]
         [:a.item {:href "https://meet.dnvgl.com/sites/open-simulation-platform-jip" :target "_blank"} [:i.icon.linkify] "JIP site"]]]]]
     [:div.ui.grid
      [:div.row
       [:div#sidebar.column
        [sidebar]]
       [:div#content.column
        [:div.ui.grid
         [:div.row
          [:h1.ui.huge.header [k/switch-route (comp :name :data)
                               :module (or @module "")
                               :trend "Trend"
                               :guide "User guide"
                               :index (if @loaded? "Simulation status" "Simulation setup")
                               nil [:div "Loading..."]]]]
         [:div.ui.divider]
         [:div.row
          [k/switch-route (comp :name :data)
           :trend [trend/trend-outer]
           :guide [guide/form]
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
          [:div.sub.header "It looks like the server is down. Try restarting the server and hit F5"]]]])
     [command-feedback-message]]))