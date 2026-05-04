"""
OctoTify API 自动化测试套件
测试范围：认证、用户、来源、渠道、消息、推送模块
覆盖：正常流程 + 异常流程 + 全流程集成测试
"""

import requests
import json
import time
import unittest
from datetime import datetime

# 配置
BASE_URL = "http://localhost:34123"
API_PREFIX = f"{BASE_URL}/api"

# 错误码定义（从后端 xerr/codes.go 同步）
ERR_CODE = {
    "BAD_REQUEST": 100000,
    "UNAUTHORIZED": 100001,
    "NOT_FOUND": 100003,
    "REGISTER_USERNAME_EXISTS": 110104,
    "LOGIN_INVALID": 110200,
    "REFRESH_TOKEN_INVALID": 110300,
    "CHANGE_PASSWORD_OLD_EMPTY": 110400,
    "CHANGE_PASSWORD_OLD_INCORRECT": 110402,
    "SOURCE_NOT_FOUND": 110503,
    "CHANNEL_INVALID_TYPE": 110601,
    "CHANNEL_NOT_FOUND": 110603,
    "MESSAGE_TITLE_EMPTY": 110700,
    "MESSAGE_CONTENT_EMPTY": 110701,
    "MESSAGE_NO_CHANNELS": 110703,
}

# 测试数据
TEST_USER = {
    "username": "test_user_001",
    "password": "Test@123456",
    "email": "test001@example.com"
}

TEST_SOURCE = {
    "name": "测试来源-自动化测试",
    "description": "自动化测试创建的来源"
}

TEST_CHANNEL = {
    "type": "webhook",
    "name": "测试渠道-Webhook",
    "config": {
        "url": "https://httpbin.org/post"
    }
}

TEST_PUSH_MESSAGE = {
    "title": "自动化测试消息",
    "message": "这是一条来自自动化测试的消息"
}


class TestResult:
    """测试结果记录器"""
    
    def __init__(self):
        self.results = []
        self.start_time = datetime.now()
    
    def add_result(self, module: str, test_name: str, status: str, 
                   expected: str, actual: str, details: str = ""):
        self.results.append({
            "module": module,
            "test_name": test_name,
            "status": status,
            "expected": expected,
            "actual": actual,
            "details": details,
            "timestamp": datetime.now().isoformat()
        })
    
    def generate_report(self) -> str:
        total = len(self.results)
        passed = sum(1 for r in self.results if r["status"] == "PASS")
        failed = sum(1 for r in self.results if r["status"] == "FAIL")
        error = sum(1 for r in self.results if r["status"] == "ERROR")
        
        report = []
        report.append("=" * 80)
        report.append("OctoTify API 自动化测试报告")
        report.append("=" * 80)
        report.append(f"测试时间: {self.start_time.strftime('%Y-%m-%d %H:%M:%S')}")
        report.append(f"测试耗时: {(datetime.now() - self.start_time).total_seconds():.2f} 秒")
        report.append(f"服务器地址: {BASE_URL}")
        report.append("")
        report.append("-" * 80)
        report.append("测试统计")
        report.append("-" * 80)
        report.append(f"总用例数: {total}")
        report.append(f"通过: {passed}")
        report.append(f"失败: {failed}")
        report.append(f"错误: {error}")
        if total > 0:
            report.append(f"通过率: {passed/total*100:.2f}%")
        else:
            report.append("通过率: 0%")
        report.append("")
        
        # 按模块分组
        modules = {}
        for r in self.results:
            if r["module"] not in modules:
                modules[r["module"]] = []
            modules[r["module"]].append(r)
        
        for module, results in modules.items():
            report.append("-" * 80)
            report.append(f"模块: {module}")
            report.append("-" * 80)
            
            module_passed = sum(1 for r in results if r["status"] == "PASS")
            module_failed = sum(1 for r in results if r["status"] == "FAIL")
            module_error = sum(1 for r in results if r["status"] == "ERROR")
            
            report.append(f"用例数: {len(results)} | 通过: {module_passed} | 失败: {module_failed} | 错误: {module_error}")
            report.append("")
            
            for r in results:
                status_icon = "✓" if r["status"] == "PASS" else "✗"
                report.append(f"  [{status_icon}] {r['test_name']}")
                report.append(f"      期望: {r['expected']}")
                report.append(f"      实际: {r['actual']}")
                if r["details"]:
                    report.append(f"      详情: {r['details']}")
                report.append("")
        
        report.append("=" * 80)
        report.append("测试结束")
        report.append("=" * 80)
        
        return "\n".join(report)


# 全局测试结果记录器
test_result = TestResult()


def check_response(resp: requests.Response, expected_code: int = 0, 
                    expected_status: int = 200) -> tuple:
    """检查响应是否符合预期"""
    if resp.status_code != expected_status:
        return False, f"HTTP状态码不匹配: 期望 {expected_status}, 实际 {resp.status_code}"
    
    try:
        data = resp.json()
        if data.get("code") != expected_code:
            return False, f"业务码不匹配: 期望 {expected_code}, 实际 {data.get('code')}, 消息: {data.get('msg')}"
        return True, "成功"
    except Exception as e:
        return False, f"解析响应失败: {str(e)}"


class TestHealthCheck(unittest.TestCase):
    """健康检查测试"""
    
    def test_ping(self):
        """测试健康检查接口"""
        resp = requests.get(f"{BASE_URL}/ping")
        passed, msg = check_response(resp, expected_code=0)
        
        if passed:
            test_result.add_result("健康检查", "ping接口", "PASS", 
                                 "code=0", f"code={resp.json().get('code')}")
        else:
            test_result.add_result("健康检查", "ping接口", "FAIL", 
                                 "code=0", msg)
        
        self.assertTrue(passed, msg)


class TestAuth(unittest.TestCase):
    """认证模块测试"""
    
    @classmethod
    def setUpClass(cls):
        """测试前清理：尝试登录"""
        resp = requests.post(f"{API_PREFIX}/auth/login", json={
            "username": TEST_USER["username"],
            "password": TEST_USER["password"]
        })
        if resp.status_code == 200 and resp.json().get("code") == 0:
            cls.access_token = resp.json()["data"]["access_token"]
            cls.refresh_token = resp.json()["data"]["refresh_token"]
        else:
            cls.access_token = None
            cls.refresh_token = None
    
    def test_01_login_success(self):
        """正常流程：用户登录成功"""
        resp = requests.post(f"{API_PREFIX}/auth/login", json={
            "username": TEST_USER["username"],
            "password": TEST_USER["password"]
        })
        
        passed, msg = check_response(resp)
        if passed:
            data = resp.json().get("data", {})
            if "access_token" in data and "refresh_token" in data:
                TestAuth.access_token = data["access_token"]
                TestAuth.refresh_token = data["refresh_token"]
                test_result.add_result("认证模块", "用户登录-成功", "PASS",
                                     "返回access_token和refresh_token", "成功获取token")
            else:
                passed = False
                msg = "响应中缺少token字段"
                test_result.add_result("认证模块", "用户登录-成功", "FAIL",
                                     "返回access_token和refresh_token", msg)
        else:
            test_result.add_result("认证模块", "用户登录-成功", "FAIL",
                                 "code=0", msg)
        
        self.assertTrue(passed, msg)
    
    def test_02_login_wrong_password(self):
        """异常流程：密码错误"""
        resp = requests.post(f"{API_PREFIX}/auth/login", json={
            "username": TEST_USER["username"],
            "password": "WrongPassword123!"
        })
        
        passed, msg = check_response(resp, expected_code=ERR_CODE["LOGIN_INVALID"])
        test_result.add_result("认证模块", "用户登录-密码错误", "PASS" if passed else "FAIL",
                             f"code={ERR_CODE['LOGIN_INVALID']}", msg)
    
    def test_03_login_user_not_found(self):
        """异常流程：用户不存在"""
        resp = requests.post(f"{API_PREFIX}/auth/login", json={
            "username": "nonexistent_user",
            "password": "SomePassword123!"
        })
        
        passed, msg = check_response(resp, expected_code=ERR_CODE["LOGIN_INVALID"])
        test_result.add_result("认证模块", "用户登录-用户不存在", "PASS" if passed else "FAIL",
                             f"code={ERR_CODE['LOGIN_INVALID']}", msg)
    
    def test_04_login_empty_fields(self):
        """异常流程：空字段"""
        resp = requests.post(f"{API_PREFIX}/auth/login", json={
            "username": "",
            "password": ""
        })
        
        passed, msg = check_response(resp, expected_code=ERR_CODE["LOGIN_INVALID"])
        test_result.add_result("认证模块", "用户登录-空字段", "PASS" if passed else "FAIL",
                             f"code={ERR_CODE['LOGIN_INVALID']}", msg)
    
    def test_05_refresh_token_success(self):
        """正常流程：刷新Token"""
        if not TestAuth.refresh_token:
            test_result.add_result("认证模块", "刷新Token-成功", "ERROR",
                                 "需要有效的refresh_token", "refresh_token不存在")
            self.skipTest("refresh_token不存在")
        
        resp = requests.post(f"{API_PREFIX}/auth/refresh", json={
            "refresh_token": TestAuth.refresh_token
        })
        
        passed, msg = check_response(resp)
        test_result.add_result("认证模块", "刷新Token-成功", "PASS" if passed else "FAIL",
                             "返回新的access_token", msg)
    
    def test_06_refresh_token_invalid(self):
        """异常流程：使用无效的refresh_token"""
        resp = requests.post(f"{API_PREFIX}/auth/refresh", json={
            "refresh_token": "invalid_token_here"
        })
        
        passed, msg = check_response(resp, expected_code=ERR_CODE["REFRESH_TOKEN_INVALID"])
        test_result.add_result("认证模块", "刷新Token-无效token", "PASS" if passed else "FAIL",
                             f"code={ERR_CODE['REFRESH_TOKEN_INVALID']}", msg)
    
    def test_07_logout_success(self):
        """正常流程：退出登录"""
        if not TestAuth.access_token:
            test_result.add_result("认证模块", "退出登录-成功", "ERROR",
                                 "需要有效的access_token", "access_token不存在")
            self.skipTest("access_token不存在")
        
        headers = {"Authorization": f"Bearer {TestAuth.access_token}"}
        resp = requests.post(f"{API_PREFIX}/auth/logout", headers=headers)
        
        passed, msg = check_response(resp)
        test_result.add_result("认证模块", "退出登录-成功", "PASS" if passed else "FAIL",
                             "code=0", msg)


class TestUser(unittest.TestCase):
    """用户模块测试"""
    
    def test_01_register_success(self):
        """正常流程：用户注册"""
        unique_user = {
            "username": f"test_user_{int(time.time())}",
            "password": "Test@123456",
            "email": f"test_{int(time.time())}@example.com"
        }
        
        resp = requests.post(f"{API_PREFIX}/user/register", json=unique_user)
        
        passed, msg = check_response(resp)
        test_result.add_result("用户模块", "用户注册-成功", "PASS" if passed else "FAIL",
                             "code=0", msg)
    
    def test_02_register_duplicate_username(self):
        """异常流程：重复用户名"""
        resp = requests.post(f"{API_PREFIX}/user/register", json=TEST_USER)
        
        passed, msg = check_response(resp, expected_code=ERR_CODE["REGISTER_USERNAME_EXISTS"])
        test_result.add_result("用户模块", "用户注册-重复用户名", "PASS" if passed else "FAIL",
                             f"code={ERR_CODE['REGISTER_USERNAME_EXISTS']}", msg)
    
    def test_03_register_weak_password(self):
        """异常流程：弱密码"""
        resp = requests.post(f"{API_PREFIX}/user/register", json={
            "username": f"test_user_{int(time.time())}",
            "password": "123",
            "email": f"test_{int(time.time())}@example.com"
        })
        
        passed, msg = check_response(resp, expected_code=ERR_CODE["BAD_REQUEST"])
        test_result.add_result("用户模块", "用户注册-弱密码", "PASS" if passed else "FAIL",
                             f"code={ERR_CODE['BAD_REQUEST']}", msg)
    
    def test_04_register_invalid_email(self):
        """异常流程：无效邮箱（后端不校验邮箱格式）"""
        unique_username = f"test_invalid_email_{int(time.time())}"
        resp = requests.post(f"{API_PREFIX}/user/register", json={
            "username": unique_username,
            "password": "Test@123456",
            "email": "invalid-email"
        })
        
        passed, msg = check_response(resp)
        test_result.add_result("用户模块", "用户注册-无效邮箱", "PASS" if passed else "FAIL",
                             "注册成功（后端不校验邮箱格式）", msg)
    
    def test_05_get_profile_success(self):
        """正常流程：获取用户信息"""
        if not TestAuth.access_token:
            test_result.add_result("用户模块", "获取用户信息-成功", "ERROR",
                                 "需要有效的access_token", "access_token不存在")
            self.skipTest("access_token不存在")
        
        headers = {"Authorization": f"Bearer {TestAuth.access_token}"}
        resp = requests.get(f"{API_PREFIX}/user/profile", headers=headers)
        
        passed, msg = check_response(resp)
        test_result.add_result("用户模块", "获取用户信息-成功", "PASS" if passed else "FAIL",
                             "返回用户信息", msg)
    
    def test_06_get_profile_no_token(self):
        """异常流程：未提供token"""
        resp = requests.get(f"{API_PREFIX}/user/profile")
        
        if resp.status_code == 401:
            test_result.add_result("用户模块", "获取用户信息-无token", "PASS",
                                 "HTTP 401 或 code=100001", "HTTP 401")
        else:
            passed, msg = check_response(resp, expected_code=ERR_CODE["UNAUTHORIZED"])
            test_result.add_result("用户模块", "获取用户信息-无token", "PASS" if passed else "FAIL",
                                 "HTTP 401 或 code=100001", msg)
    
    def test_07_change_password_success(self):
        """正常流程：修改密码"""
        if not TestAuth.access_token:
            test_result.add_result("用户模块", "修改密码-成功", "ERROR",
                                 "需要有效的access_token", "access_token不存在")
            self.skipTest("access_token不存在")
        
        headers = {"Authorization": f"Bearer {TestAuth.access_token}"}
        resp = requests.put(f"{API_PREFIX}/user/password", headers=headers, json={
            "old_password": TEST_USER["password"],
            "new_password": "NewTest@123456"
        })
        
        passed, msg = check_response(resp)
        test_result.add_result("用户模块", "修改密码-成功", "PASS" if passed else "FAIL",
                             "code=0", msg)
        
        # 恢复原密码
        requests.put(f"{API_PREFIX}/user/password", headers=headers, json={
            "old_password": "NewTest@123456",
            "new_password": TEST_USER["password"]
        })
    
    def test_08_change_password_wrong_old(self):
        """异常流程：旧密码错误"""
        if not TestAuth.access_token:
            test_result.add_result("用户模块", "修改密码-旧密码错误", "ERROR",
                                 "需要有效的access_token", "access_token不存在")
            self.skipTest("access_token不存在")
        
        headers = {"Authorization": f"Bearer {TestAuth.access_token}"}
        resp = requests.put(f"{API_PREFIX}/user/password", headers=headers, json={
            "old_password": "WrongOldPassword123!",
            "new_password": "NewTest@123456"
        })
        
        passed, msg = check_response(resp, expected_code=ERR_CODE["CHANGE_PASSWORD_OLD_INCORRECT"])
        test_result.add_result("用户模块", "修改密码-旧密码错误", "PASS" if passed else "FAIL",
                             f"code={ERR_CODE['CHANGE_PASSWORD_OLD_INCORRECT']}", msg)


class TestSource(unittest.TestCase):
    """来源模块测试"""
    
    @classmethod
    def setUpClass(cls):
        """登录获取token"""
        resp = requests.post(f"{API_PREFIX}/auth/login", json={
            "username": TEST_USER["username"],
            "password": TEST_USER["password"]
        })
        if resp.status_code == 200 and resp.json().get("code") == 0:
            cls.access_token = resp.json()["data"]["access_token"]
            cls.headers = {"Authorization": f"Bearer {cls.access_token}"}
            cls.source_id = None
        else:
            cls.access_token = None
            cls.headers = {}
            cls.source_id = None
    
    def test_01_create_source_success(self):
        """正常流程：创建来源"""
        if not self.access_token:
            test_result.add_result("来源模块", "创建来源-成功", "ERROR",
                                 "需要有效的access_token", "access_token不存在")
            self.skipTest("access_token不存在")
        
        resp = requests.post(f"{API_PREFIX}/sources", headers=self.headers, json=TEST_SOURCE)
        
        passed, msg = check_response(resp)
        if passed:
            TestSource.source_id = resp.json()["data"]["id"]
            test_result.add_result("来源模块", "创建来源-成功", "PASS",
                                 "返回来源信息", f"来源ID: {TestSource.source_id}")
        else:
            test_result.add_result("来源模块", "创建来源-成功", "FAIL",
                                 "code=0", msg)
        
        self.assertTrue(passed, msg)
    
    def test_02_create_source_empty_name(self):
        """异常流程：名称为空"""
        if not self.access_token:
            self.skipTest("access_token不存在")
        
        resp = requests.post(f"{API_PREFIX}/sources", headers=self.headers, json={
            "name": "",
            "description": "测试"
        })
        
        passed, msg = check_response(resp, expected_code=ERR_CODE["BAD_REQUEST"])
        test_result.add_result("来源模块", "创建来源-名称为空", "PASS" if passed else "FAIL",
                             f"code={ERR_CODE['BAD_REQUEST']}", msg)
    
    def test_03_get_source_list(self):
        """正常流程：获取来源列表"""
        if not self.access_token:
            self.skipTest("access_token不存在")
        
        resp = requests.get(f"{API_PREFIX}/sources", headers=self.headers)
        
        passed, msg = check_response(resp)
        test_result.add_result("来源模块", "获取来源列表", "PASS" if passed else "FAIL",
                             "返回来源列表", msg)
    
    def test_04_get_source_detail(self):
        """正常流程：获取来源详情"""
        if not self.access_token or not TestSource.source_id:
            self.skipTest("需要access_token和source_id")
        
        resp = requests.get(f"{API_PREFIX}/sources/{TestSource.source_id}", headers=self.headers)
        
        passed, msg = check_response(resp)
        test_result.add_result("来源模块", "获取来源详情", "PASS" if passed else "FAIL",
                             "返回来源详情", msg)
    
    def test_05_get_source_detail_not_found(self):
        """异常流程：来源不存在"""
        if not self.access_token:
            self.skipTest("access_token不存在")
        
        resp = requests.get(f"{API_PREFIX}/sources/999999", headers=self.headers)
        
        passed, msg = check_response(resp, expected_code=ERR_CODE["SOURCE_NOT_FOUND"])
        test_result.add_result("来源模块", "获取来源详情-不存在", "PASS" if passed else "FAIL",
                             f"code={ERR_CODE['SOURCE_NOT_FOUND']}", msg)
    
    def test_06_update_source_success(self):
        """正常流程：更新来源"""
        if not self.access_token or not TestSource.source_id:
            self.skipTest("需要access_token和source_id")
        
        resp = requests.put(f"{API_PREFIX}/sources/{TestSource.source_id}", 
                           headers=self.headers, json={
                               "name": "更新后的来源名称",
                               "description": "更新后的描述"
                           })
        
        passed, msg = check_response(resp)
        test_result.add_result("来源模块", "更新来源-成功", "PASS" if passed else "FAIL",
                             "code=0", msg)
    
    def test_07_get_source_token(self):
        """正常流程：获取来源令牌"""
        if not self.access_token or not TestSource.source_id:
            self.skipTest("需要access_token和source_id")
        
        resp = requests.post(f"{API_PREFIX}/sources/{TestSource.source_id}/token", 
                            headers=self.headers, json={
                                "password": TEST_USER["password"]
                            })
        
        passed, msg = check_response(resp)
        test_result.add_result("来源模块", "获取来源令牌", "PASS" if passed else "FAIL",
                             "返回令牌信息", msg)
    
    def test_08_get_source_token_wrong_password(self):
        """异常流程：获取令牌时密码错误"""
        if not self.access_token or not TestSource.source_id:
            self.skipTest("需要access_token和source_id")
        
        resp = requests.post(f"{API_PREFIX}/sources/{TestSource.source_id}/token", 
                            headers=self.headers, json={
                                "password": "WrongPassword123!"
                            })
        
        passed, msg = check_response(resp, expected_code=ERR_CODE["UNAUTHORIZED"])
        test_result.add_result("来源模块", "获取来源令牌-密码错误", "PASS" if passed else "FAIL",
                             f"code={ERR_CODE['UNAUTHORIZED']}", msg)
    
    def test_09_reset_source_token(self):
        """正常流程：重置来源令牌"""
        if not self.access_token or not TestSource.source_id:
            self.skipTest("需要access_token和source_id")
        
        resp = requests.put(f"{API_PREFIX}/sources/{TestSource.source_id}/token", 
                           headers=self.headers, json={
                               "password": TEST_USER["password"]
                           })
        
        passed, msg = check_response(resp)
        test_result.add_result("来源模块", "重置来源令牌", "PASS" if passed else "FAIL",
                             "返回新令牌", msg)
    
    def test_10_disable_source(self):
        """正常流程：停用来源"""
        if not self.access_token or not TestSource.source_id:
            self.skipTest("需要access_token和source_id")
        
        resp = requests.put(f"{API_PREFIX}/sources/{TestSource.source_id}/disable", 
                           headers=self.headers, json={
                               "password": TEST_USER["password"]
                           })
        
        passed, msg = check_response(resp)
        test_result.add_result("来源模块", "停用来源", "PASS" if passed else "FAIL",
                             "code=0", msg)
    
    def test_11_enable_source(self):
        """正常流程：启用来源"""
        if not self.access_token or not TestSource.source_id:
            self.skipTest("需要access_token和source_id")
        
        resp = requests.put(f"{API_PREFIX}/sources/{TestSource.source_id}/enable", 
                           headers=self.headers, json={
                               "password": TEST_USER["password"]
                           })
        
        passed, msg = check_response(resp)
        test_result.add_result("来源模块", "启用来源", "PASS" if passed else "FAIL",
                             "code=0", msg)
    
    def test_12_delete_source(self):
        """正常流程：删除来源"""
        if not self.access_token or not TestSource.source_id:
            self.skipTest("需要access_token和source_id")
        
        resp = requests.delete(f"{API_PREFIX}/sources/{TestSource.source_id}", 
                              headers=self.headers, json={
                                  "password": TEST_USER["password"]
                              })
        
        passed, msg = check_response(resp)
        test_result.add_result("来源模块", "删除来源", "PASS" if passed else "FAIL",
                             "code=0", msg)
    
    def test_13_source_no_auth(self):
        """异常流程：未认证访问"""
        resp = requests.get(f"{API_PREFIX}/sources")
        
        # 后端可能返回HTTP 401或HTTP 200 + error code
        if resp.status_code == 401:
            test_result.add_result("来源模块", "未认证访问", "PASS",
                                 "HTTP 401 或 code=100001", "HTTP 401")
        else:
            passed, msg = check_response(resp, expected_code=ERR_CODE["UNAUTHORIZED"])
            test_result.add_result("来源模块", "未认证访问", "PASS" if passed else "FAIL",
                                 "HTTP 401 或 code=100001", msg)


class TestChannel(unittest.TestCase):
    """渠道模块测试"""
    
    @classmethod
    def setUpClass(cls):
        """登录获取token"""
        resp = requests.post(f"{API_PREFIX}/auth/login", json={
            "username": TEST_USER["username"],
            "password": TEST_USER["password"]
        })
        if resp.status_code == 200 and resp.json().get("code") == 0:
            cls.access_token = resp.json()["data"]["access_token"]
            cls.headers = {"Authorization": f"Bearer {cls.access_token}"}
            cls.channel_id = None
        else:
            cls.access_token = None
            cls.headers = {}
            cls.channel_id = None
    
    def test_01_create_channel_success(self):
        """正常流程：创建渠道"""
        if not self.access_token:
            test_result.add_result("渠道模块", "创建渠道-成功", "ERROR",
                                 "需要有效的access_token", "access_token不存在")
            self.skipTest("access_token不存在")
        
        resp = requests.post(f"{API_PREFIX}/channels", headers=self.headers, json=TEST_CHANNEL)
        
        passed, msg = check_response(resp)
        if passed:
            TestChannel.channel_id = resp.json()["data"]["id"]
            test_result.add_result("渠道模块", "创建渠道-成功", "PASS",
                                 "返回渠道信息", f"渠道ID: {TestChannel.channel_id}")
        else:
            test_result.add_result("渠道模块", "创建渠道-成功", "FAIL",
                                 "code=0", msg)
        
        self.assertTrue(passed, msg)
    
    def test_02_create_channel_invalid_type(self):
        """异常流程：无效的渠道类型"""
        if not self.access_token:
            self.skipTest("access_token不存在")
        
        resp = requests.post(f"{API_PREFIX}/channels", headers=self.headers, json={
            "type": "invalid_type",
            "name": "测试渠道",
            "config": {"url": "https://example.com"}
        })
        
        passed, msg = check_response(resp, expected_code=ERR_CODE["BAD_REQUEST"])
        test_result.add_result("渠道模块", "创建渠道-无效类型", "PASS" if passed else "FAIL",
                             f"code={ERR_CODE['BAD_REQUEST']}", msg)
    
    def test_03_get_channel_list(self):
        """正常流程：获取渠道列表"""
        if not self.access_token:
            self.skipTest("access_token不存在")
        
        resp = requests.get(f"{API_PREFIX}/channels", headers=self.headers)
        
        passed, msg = check_response(resp)
        test_result.add_result("渠道模块", "获取渠道列表", "PASS" if passed else "FAIL",
                             "返回渠道列表", msg)
    
    def test_04_get_channel_detail(self):
        """正常流程：获取渠道详情"""
        if not self.access_token or not TestChannel.channel_id:
            self.skipTest("需要access_token和channel_id")
        
        resp = requests.get(f"{API_PREFIX}/channels/{TestChannel.channel_id}", headers=self.headers)
        
        passed, msg = check_response(resp)
        test_result.add_result("渠道模块", "获取渠道详情", "PASS" if passed else "FAIL",
                             "返回渠道详情", msg)
    
    def test_05_update_channel(self):
        """正常流程：更新渠道"""
        if not self.access_token or not TestChannel.channel_id:
            self.skipTest("需要access_token和channel_id")
        
        resp = requests.put(f"{API_PREFIX}/channels/{TestChannel.channel_id}", 
                           headers=self.headers, json={
                               "name": "更新后的渠道名称",
                               "config": {"url": "https://httpbin.org/post"}
                           })
        
        passed, msg = check_response(resp)
        test_result.add_result("渠道模块", "更新渠道", "PASS" if passed else "FAIL",
                             "code=0", msg)
    
    def test_06_test_channel(self):
        """正常流程：测试渠道连接"""
        if not self.access_token or not TestChannel.channel_id:
            self.skipTest("需要access_token和channel_id")
        
        resp = requests.post(f"{API_PREFIX}/channels/{TestChannel.channel_id}/test", 
                            headers=self.headers)
        
        passed, msg = check_response(resp)
        test_result.add_result("渠道模块", "测试渠道连接", "PASS" if passed else "FAIL",
                             "返回测试结果", msg)
    
    def test_07_disable_channel(self):
        """正常流程：停用渠道"""
        if not self.access_token or not TestChannel.channel_id:
            self.skipTest("需要access_token和channel_id")
        
        resp = requests.put(f"{API_PREFIX}/channels/{TestChannel.channel_id}/disable", 
                           headers=self.headers)
        
        passed, msg = check_response(resp)
        test_result.add_result("渠道模块", "停用渠道", "PASS" if passed else "FAIL",
                             "code=0", msg)
    
    def test_08_enable_channel(self):
        """正常流程：启用渠道"""
        if not self.access_token or not TestChannel.channel_id:
            self.skipTest("需要access_token和channel_id")
        
        resp = requests.put(f"{API_PREFIX}/channels/{TestChannel.channel_id}/enable", 
                           headers=self.headers)
        
        passed, msg = check_response(resp)
        test_result.add_result("渠道模块", "启用渠道", "PASS" if passed else "FAIL",
                             "code=0", msg)
    
    def test_09_delete_channel(self):
        """正常流程：删除渠道"""
        if not self.access_token or not TestChannel.channel_id:
            self.skipTest("需要access_token和channel_id")
        
        resp = requests.delete(f"{API_PREFIX}/channels/{TestChannel.channel_id}", 
                              headers=self.headers)
        
        passed, msg = check_response(resp)
        test_result.add_result("渠道模块", "删除渠道", "PASS" if passed else "FAIL",
                             "code=0", msg)
    
    def test_10_channel_no_auth(self):
        """异常流程：未认证访问"""
        resp = requests.get(f"{API_PREFIX}/channels")
        
        if resp.status_code == 401:
            test_result.add_result("渠道模块", "未认证访问", "PASS",
                                 "HTTP 401 或 code=100001", "HTTP 401")
        else:
            passed, msg = check_response(resp, expected_code=ERR_CODE["UNAUTHORIZED"])
            test_result.add_result("渠道模块", "未认证访问", "PASS" if passed else "FAIL",
                                 "HTTP 401 或 code=100001", msg)


class TestMessage(unittest.TestCase):
    """消息模块测试"""
    
    @classmethod
    def setUpClass(cls):
        """登录获取token"""
        resp = requests.post(f"{API_PREFIX}/auth/login", json={
            "username": TEST_USER["username"],
            "password": TEST_USER["password"]
        })
        if resp.status_code == 200 and resp.json().get("code") == 0:
            cls.access_token = resp.json()["data"]["access_token"]
            cls.headers = {"Authorization": f"Bearer {cls.access_token}"}
            cls.message_id = None
        else:
            cls.access_token = None
            cls.headers = {}
            cls.message_id = None
    
    def test_01_get_message_list(self):
        """正常流程：获取消息列表"""
        if not self.access_token:
            test_result.add_result("消息模块", "获取消息列表", "ERROR",
                                 "需要有效的access_token", "access_token不存在")
            self.skipTest("access_token不存在")
        
        resp = requests.get(f"{API_PREFIX}/messages", headers=self.headers)
        
        passed, msg = check_response(resp)
        test_result.add_result("消息模块", "获取消息列表", "PASS" if passed else "FAIL",
                             "返回消息列表", msg)
    
    def test_02_get_message_list_pagination(self):
        """正常流程：消息列表分页"""
        if not self.access_token:
            self.skipTest("access_token不存在")
        
        resp = requests.get(f"{API_PREFIX}/messages", headers=self.headers, params={
            "page": 1,
            "page_size": 10
        })
        
        passed, msg = check_response(resp)
        test_result.add_result("消息模块", "消息列表分页", "PASS" if passed else "FAIL",
                             "返回分页数据", msg)
    
    def test_03_filter_messages(self):
        """正常流程：筛选消息"""
        if not self.access_token:
            self.skipTest("access_token不存在")
        
        resp = requests.get(f"{API_PREFIX}/messages/filter", headers=self.headers, params={
            "page": 1,
            "page_size": 10
        })
        
        passed, msg = check_response(resp)
        test_result.add_result("消息模块", "筛选消息", "PASS" if passed else "FAIL",
                             "返回筛选结果", msg)
    
    def test_04_filter_messages_by_status(self):
        """正常流程：按状态筛选消息"""
        if not self.access_token:
            self.skipTest("access_token不存在")
        
        resp = requests.get(f"{API_PREFIX}/messages/filter", headers=self.headers, params={
            "status": 200
        })
        
        passed, msg = check_response(resp)
        test_result.add_result("消息模块", "按状态筛选消息", "PASS" if passed else "FAIL",
                             "返回筛选结果", msg)
    
    def test_05_filter_messages_by_keyword(self):
        """正常流程：关键词搜索消息"""
        if not self.access_token:
            self.skipTest("access_token不存在")
        
        resp = requests.get(f"{API_PREFIX}/messages/filter", headers=self.headers, params={
            "keyword": "测试"
        })
        
        passed, msg = check_response(resp)
        test_result.add_result("消息模块", "关键词搜索消息", "PASS" if passed else "FAIL",
                             "返回搜索结果", msg)
    
    def test_06_get_message_detail_not_found(self):
        """异常流程：消息不存在"""
        if not self.access_token:
            self.skipTest("access_token不存在")
        
        resp = requests.get(f"{API_PREFIX}/messages/999999", headers=self.headers)
        
        passed, msg = check_response(resp, expected_code=ERR_CODE["NOT_FOUND"])
        test_result.add_result("消息模块", "获取消息详情-不存在", "PASS" if passed else "FAIL",
                             f"code={ERR_CODE['NOT_FOUND']}", msg)
    
    def test_07_delete_message_not_found(self):
        """异常流程：删除不存在的消息"""
        if not self.access_token:
            self.skipTest("access_token不存在")
        
        resp = requests.delete(f"{API_PREFIX}/messages/999999", headers=self.headers)
        
        passed, msg = check_response(resp, expected_code=ERR_CODE["NOT_FOUND"])
        test_result.add_result("消息模块", "删除消息-不存在", "PASS" if passed else "FAIL",
                             f"code={ERR_CODE['NOT_FOUND']}", msg)
    
    def test_08_message_no_auth(self):
        """异常流程：未认证访问"""
        resp = requests.get(f"{API_PREFIX}/messages")
        
        if resp.status_code == 401:
            test_result.add_result("消息模块", "未认证访问", "PASS",
                                 "HTTP 401 或 code=100001", "HTTP 401")
        else:
            passed, msg = check_response(resp, expected_code=ERR_CODE["UNAUTHORIZED"])
            test_result.add_result("消息模块", "未认证访问", "PASS" if passed else "FAIL",
                                 "HTTP 401 或 code=100001", msg)


class TestPush(unittest.TestCase):
    """推送模块测试"""
    
    @classmethod
    def setUpClass(cls):
        """准备测试数据：创建来源和渠道"""
        # 登录
        resp = requests.post(f"{API_PREFIX}/auth/login", json={
            "username": TEST_USER["username"],
            "password": TEST_USER["password"]
        })
        if resp.status_code == 200 and resp.json().get("code") == 0:
            cls.access_token = resp.json()["data"]["access_token"]
            cls.headers = {"Authorization": f"Bearer {cls.access_token}"}
            
            # 创建来源
            resp = requests.post(f"{API_PREFIX}/sources", headers=cls.headers, json=TEST_SOURCE)
            if resp.status_code == 200 and resp.json().get("code") == 0:
                cls.source_id = resp.json()["data"]["id"]
                
                # 获取来源令牌
                resp = requests.post(f"{API_PREFIX}/sources/{cls.source_id}/token", 
                                    headers=cls.headers, json={
                                        "password": TEST_USER["password"]
                                    })
                if resp.status_code == 200 and resp.json().get("code") == 0:
                    cls.source_token = resp.json()["data"]["token"]
                else:
                    cls.source_token = None
            else:
                cls.source_id = None
                cls.source_token = None
            
            # 创建渠道
            resp = requests.post(f"{API_PREFIX}/channels", headers=cls.headers, json=TEST_CHANNEL)
            if resp.status_code == 200 and resp.json().get("code") == 0:
                cls.channel_id = resp.json()["data"]["id"]
            else:
                cls.channel_id = None
        else:
            cls.access_token = None
            cls.headers = {}
            cls.source_id = None
            cls.source_token = None
            cls.channel_id = None
    
    def test_01_push_message_no_channels(self):
        """正常流程：推送消息-来源未绑定渠道"""
        if not self.source_token:
            test_result.add_result("推送模块", "推送消息-来源未绑定渠道", "ERROR",
                                 "需要有效的source_token", "source_token不存在")
            self.skipTest("source_token不存在")
        
        resp = requests.post(f"{API_PREFIX}/push/{self.source_token}", json=TEST_PUSH_MESSAGE)
        
        passed, msg = check_response(resp, expected_code=ERR_CODE["MESSAGE_NO_CHANNELS"])
        test_result.add_result("推送模块", "推送消息-来源未绑定渠道", "PASS" if passed else "FAIL",
                             f"code={ERR_CODE['MESSAGE_NO_CHANNELS']}", msg)
    
    def test_02_push_message_invalid_token(self):
        """异常流程：无效的source_token"""
        resp = requests.post(f"{API_PREFIX}/push/invalid_token_here", json=TEST_PUSH_MESSAGE)
        
        passed, msg = check_response(resp, expected_code=ERR_CODE["SOURCE_NOT_FOUND"])
        test_result.add_result("推送模块", "推送消息-无效token", "PASS" if passed else "FAIL",
                             f"code={ERR_CODE['SOURCE_NOT_FOUND']}", msg)
    
    def test_03_push_message_empty_title(self):
        """异常流程：空标题"""
        if not self.source_token:
            self.skipTest("source_token不存在")
        
        resp = requests.post(f"{API_PREFIX}/push/{self.source_token}", json={
            "title": "",
            "message": "测试消息"
        })
        
        passed, msg = check_response(resp, expected_code=ERR_CODE["BAD_REQUEST"])
        test_result.add_result("推送模块", "推送消息-空标题", "PASS" if passed else "FAIL",
                             f"code={ERR_CODE['BAD_REQUEST']}", msg)
    
    def test_04_push_message_empty_content(self):
        """异常流程：空内容"""
        if not self.source_token:
            self.skipTest("source_token不存在")
        
        resp = requests.post(f"{API_PREFIX}/push/{self.source_token}", json={
            "title": "测试标题",
            "message": ""
        })
        
        passed, msg = check_response(resp, expected_code=ERR_CODE["BAD_REQUEST"])
        test_result.add_result("推送模块", "推送消息-空内容", "PASS" if passed else "FAIL",
                             f"code={ERR_CODE['BAD_REQUEST']}", msg)
    
    def test_05_push_message_missing_fields(self):
        """异常流程：缺少必填字段"""
        if not self.source_token:
            self.skipTest("source_token不存在")
        
        resp = requests.post(f"{API_PREFIX}/push/{self.source_token}", json={})
        
        passed, msg = check_response(resp, expected_code=ERR_CODE["BAD_REQUEST"])
        test_result.add_result("推送模块", "推送消息-缺少字段", "PASS" if passed else "FAIL",
                             f"code={ERR_CODE['BAD_REQUEST']}", msg)


class TestFullWorkflow(unittest.TestCase):
    """全流程集成测试"""
    
    def test_complete_workflow(self):
        """完整业务流程测试"""
        workflow_steps = []
        
        # 1. 注册新用户
        unique_user = {
            "username": f"workflow_user_{int(time.time())}",
            "password": "Workflow@123456",
            "email": f"workflow_{int(time.time())}@example.com"
        }
        resp = requests.post(f"{API_PREFIX}/user/register", json=unique_user)
        workflow_steps.append(("注册用户", resp))
        
        # 2. 登录
        resp = requests.post(f"{API_PREFIX}/auth/login", json={
            "username": unique_user["username"],
            "password": unique_user["password"]
        })
        workflow_steps.append(("登录", resp))
        
        if resp.status_code != 200 or resp.json().get("code") != 0:
            test_result.add_result("全流程测试", "完整业务流程", "FAIL",
                                 "登录成功", "登录失败")
            return
        
        access_token = resp.json()["data"]["access_token"]
        refresh_token = resp.json()["data"]["refresh_token"]
        headers = {"Authorization": f"Bearer {access_token}"}
        
        # 3. 获取用户信息
        resp = requests.get(f"{API_PREFIX}/user/profile", headers=headers)
        workflow_steps.append(("获取用户信息", resp))
        
        # 4. 创建来源
        resp = requests.post(f"{API_PREFIX}/sources", headers=headers, json={
            "name": "全流程测试来源",
            "description": "全流程测试"
        })
        workflow_steps.append(("创建来源", resp))
        
        if resp.status_code != 200 or resp.json().get("code") != 0:
            test_result.add_result("全流程测试", "完整业务流程", "FAIL",
                                 "创建来源成功", "创建来源失败")
            return
        
        source_id = resp.json()["data"]["id"]
        
        # 5. 获取来源令牌
        resp = requests.post(f"{API_PREFIX}/sources/{source_id}/token", 
                            headers=headers, json={
                                "password": unique_user["password"]
                            })
        workflow_steps.append(("获取来源令牌", resp))
        
        if resp.status_code != 200 or resp.json().get("code") != 0:
            test_result.add_result("全流程测试", "完整业务流程", "FAIL",
                                 "获取令牌成功", "获取令牌失败")
            return
        
        source_token = resp.json()["data"]["token"]
        
        # 6. 创建渠道
        resp = requests.post(f"{API_PREFIX}/channels", headers=headers, json={
            "type": "webhook",
            "name": "全流程测试渠道",
            "config": {"url": "https://httpbin.org/post"}
        })
        workflow_steps.append(("创建渠道", resp))
        
        # 7. 推送消息（预期失败，因为来源未绑定渠道）
        resp = requests.post(f"{API_PREFIX}/push/{source_token}", json={
            "title": "全流程测试消息",
            "message": "这是一条全流程测试消息"
        })
        workflow_steps.append(("推送消息", resp))
        
        # 8. 查看消息列表
        resp = requests.get(f"{API_PREFIX}/messages", headers=headers)
        workflow_steps.append(("查看消息列表", resp))
        
        # 9. 刷新Token
        resp = requests.post(f"{API_PREFIX}/auth/refresh", json={
            "refresh_token": refresh_token
        })
        workflow_steps.append(("刷新Token", resp))
        
        # 10. 退出登录
        resp = requests.post(f"{API_PREFIX}/auth/logout", headers=headers)
        workflow_steps.append(("退出登录", resp))
        
        # 评估结果（推送消息预期会失败，因为来源未绑定渠道）
        expected_success_steps = ["注册用户", "登录", "获取用户信息", "创建来源", "获取来源令牌", "创建渠道", "查看消息列表", "刷新Token", "退出登录"]
        all_passed = all(
            step[1].status_code == 200 and step[1].json().get("code") == 0
            for step in workflow_steps if step[0] in expected_success_steps
        )
        
        push_resp = [step[1] for step in workflow_steps if step[0] == "推送消息"][0]
        push_expected_fail = push_resp.status_code == 200 and push_resp.json().get("code") == 110703
        
        if all_passed and push_expected_fail:
            test_result.add_result("全流程测试", "完整业务流程", "PASS",
                                 "核心流程成功，推送按预期失败", "全流程测试通过（推送未绑定渠道）")
        else:
            failed_steps = [
                step[0] for step in workflow_steps 
                if (step[0] in expected_success_steps and (step[1].status_code != 200 or step[1].json().get("code") != 0))
            ]
            if not push_expected_fail:
                failed_steps.append("推送消息(预期失败)")
            test_result.add_result("全流程测试", "完整业务流程", "FAIL",
                                 "核心流程成功，推送按预期失败", f"失败的步骤: {', '.join(failed_steps)}")


def run_tests():
    """运行所有测试并生成报告"""
    print("=" * 80)
    print("开始运行 OctoTify API 自动化测试")
    print("=" * 80)
    print()
    
    # 创建测试套件
    loader = unittest.TestLoader()
    suite = unittest.TestSuite()
    
    # 添加测试类
    suite.addTests(loader.loadTestsFromTestCase(TestHealthCheck))
    suite.addTests(loader.loadTestsFromTestCase(TestAuth))
    suite.addTests(loader.loadTestsFromTestCase(TestUser))
    suite.addTests(loader.loadTestsFromTestCase(TestSource))
    suite.addTests(loader.loadTestsFromTestCase(TestChannel))
    suite.addTests(loader.loadTestsFromTestCase(TestMessage))
    suite.addTests(loader.loadTestsFromTestCase(TestPush))
    suite.addTests(loader.loadTestsFromTestCase(TestFullWorkflow))
    
    # 运行测试
    runner = unittest.TextTestRunner(verbosity=2)
    result = runner.run(suite)
    
    # 生成报告
    print()
    report = test_result.generate_report()
    print(report)
    
    # 保存报告到文件
    with open("/tmp/octotify_test_report.txt", "w", encoding="utf-8") as f:
        f.write(report)
    
    print()
    print("测试报告已保存到: /tmp/octotify_test_report.txt")
    
    return result.wasSuccessful()


if __name__ == "__main__":
    success = run_tests()
    exit(0 if success else 1)
