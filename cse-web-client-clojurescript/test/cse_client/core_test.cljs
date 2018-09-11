(ns cse-client.core-test
  (:require
    [cljs.test :refer-macros [deftest is testing]]
    [cse-client.core :as core]))

(deftest multiply-test
  (is (not= 1 core/routes)))

