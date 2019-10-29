(ns cse-client.controller
  (:require-macros
    [cljs.core.async.macros :refer [go go-loop]])
  (:require [kee-frame.core :as k]
            [kee-frame.websocket :as websocket]
            [cse-client.config :refer [socket-url]]
            [re-frame.loggers :as re-frame-log]
            [cljs.spec.alpha :as s]
            [cljs.core.async :refer [<! timeout]]
            [clojure.string :as str]
            [re-frame.core :as rf]
            [cse-client.localstorage :as storage]
            [cse-client.msgpack-format]))

;; Prevent handler overwriting warnings during cljs reload.
(re-frame-log/set-loggers!
  {:warn (fn [& args]
           (when-not (or (re-find #"^re-frame: overwriting" (first args))
                         (re-find #"^Overwriting controller" (first args)))
             (apply js/console.warn args)))})

(k/reg-controller :module
                  {:params (fn [route]
                             (when (-> route :data :name (= :module))
                               (-> route :path-params)))
                   :start  [::module-enter]
                   :stop   [::module-leave]})

(k/reg-controller :trend
                  {:params (fn [route]
                             (when (-> route :data :name (= :trend))
                               (-> route :path-params)))
                   :start  [::trend-enter]
                   :stop   [::trend-leave]})

(k/reg-controller :scenario
                  {:params (fn [route]
                             (when (-> route :data :name (= :scenario))
                               (-> route :path-params)))
                   :start  [::scenario-enter]
                   :stop   [::scenario-leave]})

(k/reg-controller :websocket-controller
                  {:params (constantly true)
                   :start  [:start-websockets]})

(defn socket-command [cmd]
  {:dispatch [::websocket/send socket-url {:command cmd}]})

(k/reg-event-fx :start-websockets
                (fn [_ _]
                  (merge {::websocket/open {:path         socket-url
                                            :dispatch     ::socket-message-received
                                            :format       :msgpack
                                            :wrap-message identity}}
                         (socket-command ["get-module-data"]))))

(s/def ::fmu (s/keys :req-un [::name ::index ::variables]))
(s/def ::fmus (s/coll-of ::fmu))
(s/def ::module-data (s/keys :req-un [::fmus]))

(k/reg-event-fx ::feedback-message
                (fn [{:keys [db]} [message]]
                  (if-not (:success message)
                    {:db (assoc db :feedback-message message)}
                    (when (:show-success-feedback-messages db)
                      (go
                        (<! (timeout 3000))
                        (rf/dispatch [::close-feedback-message]))
                      {:db (assoc db :feedback-message message)}))))

(k/reg-event-db ::close-feedback-message
                (fn [db _]
                  (dissoc db :feedback-message)))

(k/reg-event-fx ::socket-message-received
                (fn [{:keys [db]} [{message :message}]]
                  (when-let [module-data (:module-data message)]
                    (s/assert ::module-data module-data)
                    (rf/dispatch [::fetch-signals]))
                  (merge
                    {:db (update db :state merge message)}
                    (when-let [feedback (:feedback message)]
                      {:dispatch [::feedback-message feedback]}))))

(k/reg-event-fx ::fetch-module-data
                (fn [_ _]
                  (socket-command ["get-module-data"])))

(defn encode-variable [{:keys [name causality type value-reference]}]
  (str/join "," [name causality type value-reference]))

(defn page-size [db]
  (or (:vars-per-page db)
      10))

(defn module-meta [db module]
  (->> db
       :state
       :module-data
       :fmus
       (filter #(= module (:name %)))
       first))

(defn variable-groups [module-meta causality vars-per-page]
  (->> module-meta
       :variables
       (filter #(= causality (:causality %)))
       (sort-by :name)
       (partition-all vars-per-page)))

(defn filter-signals [groups page]
  (some-> groups
          seq
          (nth (dec page))))

(defn editable? [{:keys [type causality] :as variable}]
  (if (and (#{"input" "parameter" "calculatedParameter" "output"} causality)
           (#{"Real" "Integer" "Boolean" "String"} type))
    (assoc variable :editable? true)
    variable))

(k/reg-event-db ::guide-navigate
                (fn [db [header]]
                  (assoc db :active-guide-tab header)))

(k/reg-event-fx ::fetch-signals
                (fn [{:keys [db]} _]
                  (let [{:keys [current-module active-causality page current-module-meta vars-per-page]} db
                        groups (variable-groups current-module-meta active-causality vars-per-page)
                        viewing (filter-signals groups page)]
                    (merge
                      {:db (assoc db :viewing (map editable? viewing)
                                     :page-count (count groups))}
                      (socket-command (concat ["signals" current-module] (map encode-variable viewing)))))))

(k/reg-event-fx ::module-enter
                (fn [{:keys [db]} [{:keys [module causality]}]]
                  (merge
                    {:db       (assoc db :current-module module
                                         :current-module-meta (module-meta db module)
                                         :active-causality causality
                                         :page 1)
                     :dispatch [::fetch-signals]})))

(k/reg-event-fx ::module-leave
                (fn [{:keys [db]} _]
                  (when (not= (:current-module db)
                              (get-in db [:kee-frame/route :path-params :module]))
                    (merge {:db (dissoc db :current-module :current-module-meta)}
                           (socket-command ["signals"])))))

(k/reg-event-fx ::trend-enter
                (fn [{:keys [db]} [{:keys [index]}]]
                  (let [trend-id (-> db :state :trends (get (int index)) :id)]
                    (merge
                      {:db (assoc db :active-trend-index index)}
                      (socket-command ["active-trend" (str trend-id)])))))

(k/reg-event-fx ::trend-leave
                (fn [{:keys [db]} _]
                  (merge
                    {:db (dissoc db :active-trend-index)}
                    (socket-command ["active-trend" nil]))))

(k/reg-event-db ::toggle-show-success-feedback-messages
                (fn [db _]
                  (let [new-setting (not (:show-success-feedback-messages db))]
                    (storage/set-item! "show-success-feedback-message" new-setting)
                    (assoc db :show-success-feedback-messages new-setting))))

(k/reg-event-fx ::scenario-enter
                (fn [{:keys [db]} [{:keys [id]}]]
                  (merge {:db (assoc db :scenario-id id)}
                         (socket-command ["parse-scenario" id]))))

(k/reg-event-db ::scenario-leave
                (fn [db _]
                  (dissoc db :scenario-id)))

(k/reg-event-fx ::load
                (fn [{:keys [db]} [folder log-folder]]
                  (let [paths (distinct (conj (:prev-paths db) folder))]
                    (storage/set-item! "cse-paths" (pr-str paths))
                    (merge
                      {:db (assoc db :prev-paths paths)}
                      (socket-command ["load" folder (or log-folder "")])))))

(k/reg-event-fx ::delete-prev
                (fn [{:keys [db]} [path]]
                  (let [paths (remove #(= path %) (:prev-paths db))]
                    (storage/set-item! "cse-paths" (pr-str paths))
                    {:db (assoc db :prev-paths paths)})))

(k/reg-event-fx ::teardown
                (fn [_ _]
                  (socket-command ["teardown"])))

(k/reg-event-fx ::play
                (fn [_ _]
                  (socket-command ["play"])))

(k/reg-event-fx ::pause
                (fn [_ _]
                  (socket-command ["pause"])))

(k/reg-event-fx ::enable-realtime
                (fn [_]
                  (socket-command ["enable-realtime"])))

(k/reg-event-fx ::set-real-time-factor-target
                (fn [{:keys [db]} [val]]
                  (merge
                    {:db (assoc db :enable-real-time-target true)}
                    (socket-command ["set-custom-realtime-factor" val]))))

(k/reg-event-fx ::disable-realtime
                (fn [_ _]
                  (socket-command ["disable-realtime"])))

(k/reg-event-fx ::untrend
                (fn [_ [id]]
                  (socket-command ["untrend" (str id)])))

(k/reg-event-fx ::removetrend
                (fn [{:keys [db]} [id]]
                  (let [route-name (:name (:data (:kee-frame/route db)))
                        route-param-index (int (:index (:path-params (:kee-frame/route db))))
                        current-path-to-be-deleted (and (= :trend route-name) (= route-param-index id))]
                    (merge
                      (when current-path-to-be-deleted {:navigate-to [:index]})
                      (socket-command ["removetrend" (str id)])))))

(k/reg-event-fx ::new-trend
                (fn [_ [type label]]
                  (socket-command ["newtrend" type label])))

(k/reg-event-fx ::add-to-trend
                (fn [_ [module signal causality type value-reference plot-index]]
                  (socket-command ["addtotrend" module signal causality type (str value-reference) (str plot-index)])))

(k/reg-event-fx ::set-label
                (fn [{:keys [db]} [label]]
                  (socket-command ["setlabel" (:active-trend-index db) label])))

(k/reg-event-fx ::set-value
                (fn [_ [index type value-reference value]]
                  (socket-command ["set-value" index type value-reference (str value)])))

(k/reg-event-fx ::reset-value
                (fn [_ [index type value-reference]]
                  (socket-command ["reset-value" index type value-reference])))

(k/reg-event-fx ::trend-zoom
                (fn [{:keys [db]} [begin end]]
                  (socket-command ["trend-zoom" (:active-trend-index db) (str begin) (str end)])))

(k/reg-event-fx ::trend-zoom-reset
                (fn [{:keys [db]} _]
                  (socket-command ["trend-zoom-reset" (:active-trend-index db) (-> db :trend-range str)])))

(k/reg-event-fx ::trend-range
                (fn [{:keys [db]} [new-range]]
                  {:db       (assoc db :trend-range new-range)
                   :dispatch [::trend-zoom-reset]}))

(k/reg-event-fx ::set-page
                (fn [{:keys [db]} [page]]
                  {:db       (assoc db :page page)
                   :dispatch [::fetch-signals]}))

(k/reg-event-fx ::set-vars-per-page
                (fn [{:keys [db]} [n]]
                  {:db       (assoc db :vars-per-page (max 1 n)
                                       :page 1)
                   :dispatch [::fetch-signals]}))

(k/reg-event-fx ::load-scenario
                (fn [_ [file-name]]
                  (socket-command ["load-scenario" file-name])))

(k/reg-event-fx ::abort-scenario
                (fn [_ [file-name]]
                  (socket-command ["abort-scenario" file-name])))

(k/reg-event-db ::toggle-dismiss-error
                (fn [db]
                  (update db :error-dismissed not)))
