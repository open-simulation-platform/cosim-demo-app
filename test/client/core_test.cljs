(ns client.core-test
  (:require
    [cljs.test :refer-macros [deftest is testing]]
    [client.core :as core]))

(deftest root-component-test
  (is (= :div (first (core/root-comp)))))

(deftest ws-request
  (is (= "play" (:command (core/ws-request "play")))))