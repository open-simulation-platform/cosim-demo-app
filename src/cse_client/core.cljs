(ns cse-client.core
  (:require [kee-frame.core :as k]
            [re-frame.core :as rf]
            [cse-client.view :as view]
            [cse-client.controller :as controller]))

(enable-console-print!)

(def routes
  [["/" :index]
   ["/modules/:module" :module]
   ["/trend/:module/:signal" :trend]])

(rf/reg-sub :module (comp :module :state))
(rf/reg-sub :modules (comp :modules :state))


(k/start! {:routes         routes
           :hash-routing?  true
           :debug?         {:blacklist #{::controller/socket-message-received}}
           :root-component [view/root-comp]
           :initial-db     {:trend-values []}})