;; This Source Code Form is subject to the terms of the Mozilla Public
;; License, v. 2.0. If a copy of the MPL was not distributed with this
;; file, You can obtain one at https://mozilla.org/MPL/2.0/.

(ns client.localstorage
  (:require [oops.core :refer [ocall oget]]))

(def local-storage
  (oget js/window "localStorage"))

(defn set-item!
  [key val]
  (ocall local-storage "setItem" key val))

(defn get-item
  [key]
  (ocall local-storage "getItem" key))

(defn remove-item!
  [key]
  (ocall local-storage "removeItem" key))