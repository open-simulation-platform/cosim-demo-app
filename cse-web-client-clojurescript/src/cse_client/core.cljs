(ns cse-client.core
  (:require [kee-frame.core :as k]
            [kee-frame.websocket :as websocket]
            [re-frame.core :as rf]
            [clojure.string :as string]
            [kee-frame.api :as api]
            [reitit.core :as reitit]
            [cse-client.trend :as trend]))

(def socket-url "ws://localhost:8000/ws")

(enable-console-print!)

(def routes
  [["/" :index]
   ["/modules/:name" :module]])

(defonce router (reitit/router routes))

(defrecord ReititRouter [routes]
  api/Router

  (data->url [_ [route-name path-params]]
    (str (:path (reitit/match-by-name routes route-name path-params))
         (when-some [q (:query-string path-params)] (str "?" q))
         (when-some [h (:hash path-params)] (str "#" h))))

  (url->data [_ url]
    (let [[path+query fragment] (-> url (string/replace #"^/#" "") (string/split #"#" 2))
          [path query] (string/split path+query #"\?" 2)]
      (some-> (reitit/match-by-path routes path)
              (assoc :query-string query :hash fragment)))))


(k/reg-controller :websocket-controller
                  {:params #(when (-> % :data :name (= :index)) true)
                   :start  [:start-websockets]
                   :stop   [:stop-websockets]})

(k/reg-event-fx :start-websockets
                (fn [_ _]
                  {::websocket/open {:path         socket-url
                                     :dispatch     ::socket-message-received
                                     :format       :json-kw
                                     :wrap-message identity}}))

(k/reg-event-db ::socket-message-received
                (fn [db [{message :message}]]
                  (let [value (-> message :module :signals first :value)]
                    (-> db
                        (update-in [:trend-values 0 :trend-data] conj [value value])
                        (update :state merge message)))))

(defn ws-request [command]
  (merge
    (when command
      {:command command})
    {:module      "Clock"
     :modules     false
     :connections false}))

(k/reg-event-fx :play
                (fn [_ _]
                  {:dispatch [::websocket/send socket-url (ws-request "play")]}))


(k/reg-event-fx :pause
                (fn [_ _]
                  {:dispatch [::websocket/send socket-url (ws-request "pause")]}))

(k/reg-event-db :trend
                (fn [db _]
                  (assoc db :trend-values [{:trend-data []
                                            :module     "Clock"
                                            :signal     "Clock"}])))

(rf/reg-sub :state :state)

(defn root-comp []
  (let [{:keys [module modules]} @(rf/subscribe [:state])
        {:keys [name signals]} module]
    [:div
     [:h3 "Modules"]
     [:ul
      (map (fn [module]
             [:li {:key module} [:a {:href (k/path-for [:module {:name module}])} module]])
           modules)]
     [:h3 "Selected module: " name]
     [:ul
      (map (fn [{:keys [name value]}]
             [:li {:key (str module "_ " name)} "Signal name: " name
              [:ul
               [:li "Signal value: " value]]])
           signals)]
     [:div.ui.buttons
      [:button.ui.button {:on-click #(rf/dispatch [:play])} "Play"]
      [:button.ui.button {:on-click #(rf/dispatch [:pause])} "Pause"]
      [:button.ui.button {:on-click #(rf/dispatch [:trend])} "Trend"]]
     [trend/trend]]))

(k/start! {:router         (->ReititRouter router)
           :hash-routing?  true
           :debug?         {:blacklist #{::socket-message-received}}
           :root-component [root-comp]
           :initial-db     {:trend-values []}})